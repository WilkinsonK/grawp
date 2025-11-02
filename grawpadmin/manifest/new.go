package manifest

import (
	"fmt"
	"os"
	"path/filepath"
)

type ServiceNewCallback func(opts ServiceNewOpts) error

type ServiceNewOpts struct {
	Callbacks        []ServiceNewCallback
	LocalVolume      string
	MinecraftVersion string
	ServiceName      string
	ServicePath      string
}

func (Sn *ServiceNewOpts) GetServicePath() string {
	return filepath.Join(Sn.ServicePath, Sn.ServiceName)
}

func openWritableFile(name string) (*os.File, error) {
	return os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, defaultFileMode)
}

func GenerateDockerFile(opts ServiceNewOpts) error {
	file, err := openWritableFile(filepath.Join(opts.GetServicePath(), ".Dockerfile"))
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString("# Default .Dockerfile\n")
	file.WriteString("# Define how/what the service should do here.\n")
	file.WriteString("FROM scratch\n")
	file.WriteString("WORKDIR /opt/\n")
	file.WriteString("COPY . .\n")
	file.WriteString("CMD [ \"sh\", \"-c\", \"echo 'Hello, World!'\" ]\n")

	return nil
}

func GenerateDockerIgnoreFile(opts ServiceNewOpts) error {
	file, err := openWritableFile(filepath.Join(opts.GetServicePath(), ".dockerignore"))
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString("*.Dockerfile\n")
	file.WriteString("*service.yaml\n")
	file.WriteString("*.tmpl\n")
	file.WriteString("*.template\n")

	return nil
}

func GenerateServiceFile(opts ServiceNewOpts) error {
	file, err := openWritableFile(filepath.Join(opts.GetServicePath(), "service.yaml"))
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintf(file, "# This is the name of the service. The name is used to\n")
	fmt.Fprintf(file, "# construct service images and containers\n")
	fmt.Fprintf(file, "name: %s\n", opts.ServiceName)
	fmt.Fprintf(file, "minecraft-version: %s\n", opts.MinecraftVersion)
	fmt.Fprintf(file, "# Args are used at image build-time. Any declared\n")
	fmt.Fprintf(file, "# \"ARG\" calls in the Docker file can be defined here\n")
	fmt.Fprintf(file, "args:\n")
	fmt.Fprintf(file, "# The volume mount from the host filesystem. Volume\n")
	fmt.Fprintf(file, "# mounts from the container point to /opt/.\n")
	fmt.Fprintf(file, "local-volume: %s\n", opts.LocalVolume)
	fmt.Fprintf(file, "ports:\n")
	fmt.Fprintf(file, "# Properties can be any arbitrary value. Like Args,\n")
	fmt.Fprintf(file, "# except that they are not used at build time.\n")
	fmt.Fprintf(file, "properties:\n")
	fmt.Fprintf(file, "tags:\n")
	fmt.Fprintf(file, "  - \"{{.Name}}:latest\"\n")
	fmt.Fprintf(file, "  - \"{{.Name}}:{{.MinecraftVersion}\"\n")

	return nil
}

func ServiceNew(gm GrawpManifest) error {
	return ServiceNewWithOpts(ServiceNewOpts{
		Callbacks: []ServiceNewCallback{
			GenerateDockerFile,
			GenerateDockerIgnoreFile,
			GenerateServiceFile,
		},
		MinecraftVersion: gm.metadata.MinecraftVersion,
		LocalVolume:      ".",
		ServiceName:      gm.metadata.Service.Name,
		ServicePath:      gm.ServicesPath,
	})
}

func ServiceNewWithOpts(opts ServiceNewOpts) error {
	os.Mkdir(opts.GetServicePath(), defaultFileMode)
	for _, callback := range opts.Callbacks {
		if err := callback(opts); err != nil {
			return err
		}
	}
	return nil
}
