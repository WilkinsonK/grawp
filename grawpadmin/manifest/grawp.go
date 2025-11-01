package manifest

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"github.com/goccy/go-yaml"
)

const defaultMode = 0644
const dotGrawpName = "*.grawp"
const grawpManifestName = "grawp.yaml"
const grawpManifestDefaultData = "data-name: \"data.db\"\nservices-path: \"{{.ProjectDir}}/services\""

var deadPaths []string
var foundPath string

type GrawpManifest struct {
	manifestPath string
	DataName     string `json:"data-name"`
	ServicesPath string `json:"services-path"`
	ProjectDir   string
}

func (Gm *GrawpManifest) formatString(templateName string, value string) (string, error) {
	templ, err := template.New(templateName).Parse(value)
	if err != nil {
		return "", nil
	}

	var buf bytes.Buffer
	if err = templ.Execute(&buf, Gm); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (Gm *GrawpManifest) GetManifestDirectory() string {
	return Gm.manifestPath
}

func (Gm *GrawpManifest) GetServicesPath() (string, error) {
	return Gm.formatString("ServicesPath", Gm.ServicesPath)
}

func GenerateDotGrawp() error {
	os.Mkdir(strings.ReplaceAll(dotGrawpName, "*", ""), defaultMode)
	ResetDeadPaths()
	FindDotGrawp()

	fileName := filepath.Join(foundPath, grawpManifestName)
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, defaultMode)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(grawpManifestDefaultData)
	if err != nil {
		return err
	}

	return nil
}

// Load a manifest from some buffer.
func LoadsGrawpManifest(name string, buffer []byte) (GrawpManifest, error) {
	var output GrawpManifest = GrawpManifest{}
	var err error = nil
	err = yaml.Unmarshal(buffer, &output)
	output.manifestPath = filepath.Dir(name)

	if output.ProjectDir == "" {
		output.ProjectDir = filepath.Dir(output.manifestPath)
	}

	return output, err
}

// Load a manifest from some file path.
func LoadGrawpManifest() (GrawpManifest, error) {
	var err error = nil
	var data []byte

	if !PathFound() {
		_, err = FindDotGrawp()
		if err != nil {
			err = GenerateDotGrawp()
		}
	}
	if err == nil {
		name := filepath.Join(foundPath, grawpManifestName)
		if data, err = os.ReadFile(name); err == nil {
			return LoadsGrawpManifest(name, data)
		}
	}

	return GrawpManifest{}, err
}

func AddDeadPath(path string) {
	deadPaths = append(deadPaths, path)
}

func FindDotGrawp() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return FindDotGrawpFromPaths(wd, home)
}

func FindDotGrawpFromParent(path string) (string, error) {
	return FindDotGrawpFromPath(filepath.Dir(path))
}

func FindDotGrawpFromPath(path string) (string, error) {
	if IsDeadPath(path) && !IsRootPath(path) {
		return FindDotGrawpFromParent(path)
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if yes, err := filepath.Match(dotGrawpName, file.Name()); err != nil {
			continue
		} else if !yes || file.Type().IsRegular() {
			continue
		}
		foundPath = filepath.Join(path, file.Name())
		return foundPath, nil
	}

	AddDeadPath(path)
	if !IsRootPath(path) {
		return FindDotGrawpFromParent(path)
	}
	return "", fmt.Errorf("No '%s' in %s", dotGrawpName, path)
}

func FindDotGrawpFromPaths(paths ...string) (string, error) {
	for _, path := range paths {
		if found, err := FindDotGrawpFromPath(path); err == nil {
			return found, nil
		}
	}
	return "", fmt.Errorf("No '%s' in any of %v\n", dotGrawpName, paths)
}

func IsRootPath(path string) bool {
	return path == "/"
}

func IsDeadPath(path string) bool {
	return slices.Contains(deadPaths, path)
}

func PathFound() bool {
	return foundPath != ""
}

func ResetDeadPaths() {
	deadPaths = []string{}
}

func ResetFoundPath() {
	foundPath = ""
}
