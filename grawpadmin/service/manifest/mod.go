// Defines and dictates the structure of container image
// manifestation per each server container definition.
package manifest

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/goccy/go-yaml"
	"github.com/moby/go-archive"
)

type ServiceManifest struct {
	manifestPath     string
	Name             string
	Dockerfile       string
	MinecraftVersion string
	Args             map[string]any
	LocalVolume      string
	Ports            []string
	Properties       map[string]any
	Tags             []string
}

func (Sm *ServiceManifest) formatString(templateName string, value string) (string, error) {
	templ, err := template.New(templateName).Parse(value)
	if err != nil {
		return "", nil
	}

	var buf bytes.Buffer
	if err = templ.Execute(&buf, Sm); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Get a string value from the `Arg` list.
//
// Args are capable of being formatted by manifest values
// using `{{ ... }}` contexts.
func (Sm *ServiceManifest) GetArgS(key string) (string, error) {
	value := Sm.Args[key].(string)
	return Sm.formatString(fmt.Sprintf("%s-arg-template", key), value)
}

// Get the Dockerfile name associated with this image
// manifest.
func (Sm *ServiceManifest) GetDockerfile() string {
	if Sm.Dockerfile != "" {
		return Sm.Dockerfile
	}
	return ".Dockerfile"
}

// Get the image build context associated with this image
// manifest.
func (Sm *ServiceManifest) GetImageBuildContext() (io.ReadCloser, error) {
	return archive.TarWithOptions(Sm.GetManifestDirectory(), &archive.TarOptions{})
}

// Generates options for building a container image.
func (Sm *ServiceManifest) GetImageBuildOptions() (build.ImageBuildOptions, error) {
	var err error
	opts := build.ImageBuildOptions{}
	opts.BuildArgs = make(map[string]*string)
	opts.Dockerfile = Sm.GetDockerfile()
	opts.ForceRemove = true

	if tags, err := Sm.GetTags(); err == nil {
		opts.Tags = append(opts.Tags, tags...)
	}

	for key := range Sm.Args {
		var v string
		if v, err = Sm.GetArgS(key); err != nil {
			continue
		}
		opts.BuildArgs[key] = &v
	}

	return opts, err
}

// Returns the directory from where the image manifest
// exists on the filesystem.
func (Sm *ServiceManifest) GetManifestDirectory() string {
	return filepath.Dir(Sm.manifestPath)
}

// Gets the ports as mappings that can be used to bind
// between the service container and the host.
func (Sm *ServiceManifest) GetPorts() (nat.PortSet, nat.PortMap, error) {
	return nat.ParsePortSpecs(Sm.Ports)
}

// Get a string value from the `Properties` list.
//
// Properties are capable of being formatted by manifest
// values using `{{ ... }}` contexts.
func (Sm *ServiceManifest) GetPropertyS(key string) (string, error) {
	value := Sm.Properties[key].(string)
	return Sm.formatString(fmt.Sprintf("%s-property-template", key), value)
}

// Attempts to generate a build config for creating a
// service container.
func (Sm *ServiceManifest) GetServiceBuildConfig(tagName string) (container.Config, error) {
	var imageName string = ""
	var config container.Config

	// Try to identify which tag to use.
	tags, err := Sm.GetTags()
	if err != nil {
		return config, err
	}
	for _, tag := range tags {
		if strings.Contains(tag, tagName) {
			imageName = tag
			break
		}
	}

	if imageName == "" {
		return config, fmt.Errorf("No container image available as '%s'", tagName)
	}

	// Try to get port mappings (if any).
	portSet, _, err := Sm.GetPorts()
	if err != nil {
		return config, err
	}

	config.Image = imageName
	config.ExposedPorts = portSet
	return config, nil
}

// Attempts to generate a container host config used to
// create a service container from.
func (Sm *ServiceManifest) GetServiceHostConfig() (container.HostConfig, error) {
	// Configure host to allow volume bindings between the
	// service and the local file system.
	hostConfig := container.HostConfig{}
	if Sm.LocalVolume != "" {
		if _, err := os.Stat(Sm.LocalVolume); os.IsNotExist(err) {
			err = os.MkdirAll(Sm.LocalVolume, 0755)
			if err != nil {
				return hostConfig, err
			}
		}
		hostConfig.Binds = []string{fmt.Sprintf("%s:/opt", Sm.LocalVolume)}
	}

	// Try to get port mappings (if any).
	_, portBindings, err := Sm.GetPorts()
	if err != nil {
		return hostConfig, err
	}
	hostConfig.PortBindings = portBindings

	return hostConfig, nil
}

// Get the service name to create a service container as.
func (Sm *ServiceManifest) GetServiceName() string {
	return strings.Join([]string{"service", Sm.Name, Sm.MinecraftVersion}, "-")
}

// Get the container image tag name.
//
// Tags are capable of being formatted by manifest values
// using `{{ ... }}` contexts.
func (Sm *ServiceManifest) GetTags() ([]string, error) {
	var tags []string = []string{}
	for tag := range Sm.Tags {
		t, err := Sm.formatString("tag-template", Sm.Tags[tag])
		if err != nil {
			return tags, err
		}
		tags = append(tags, t)
	}
	return tags, nil
}

// Load a manifest from some buffer.
func LoadsManifest(fileName string, buffer []byte) (ServiceManifest, error) {
	var output ServiceManifest = ServiceManifest{}
	var err error = nil
	err = yaml.Unmarshal(buffer, &output)
	output.manifestPath = fileName
	return output, err
}

// Load a manifest from some file path.
func LoadManifest(fileName string) (ServiceManifest, error) {
	var err error = nil
	var data []byte
	if data, err = os.ReadFile(fileName); err == nil {
		return LoadsManifest(fileName, data)
	}
	return ServiceManifest{}, err
}
