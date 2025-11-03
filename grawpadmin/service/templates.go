package service

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/WilkinsonK/grawp/grawpadmin/manifest"
)

const defaultFileMode = 0755

// Find all template files in a directory path.
//
// Template files are designated using the `.tmpl`
// extension. This makes it easiest to find and parse these
// files.
func FindTemplateFiles(path string) ([]string, error) {
	var templateFiles []string
	var err error
	files, err := os.ReadDir(path)
	if err != nil {
		return templateFiles, err
	}
	for _, file := range files {
		if !(file.Type().IsRegular() && strings.HasSuffix(file.Name(), ".tmpl")) {
			continue
		}
		templateFiles = append(templateFiles, filepath.Join(path, file.Name()))
	}

	return templateFiles, nil
}

// Find templates that belong to a specified
// `ServiceManifest`.
func FindTemplateFilesFromManifest(sm manifest.ServiceManifest) ([]string, error) {
	return FindTemplateFiles(sm.GetManifestDirectory())
}

// Load a single template from its file name.
func LoadTemplate(templateName string) (*template.Template, error) {
	data, err := os.ReadFile(templateName)
	if err != nil {
		return nil, err
	}
	tmpl, err := template.New(filepath.Base(templateName)).Parse(string(data))
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

// Acquire and render all file assets from service
// templates.
func RenderAllFromManifest(sm *manifest.ServiceManifest) error {
	root := sm.GetAssetsDirectory()
	os.MkdirAll(root, defaultFileMode)
	files, err := sm.GetTemplateFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		into := filepath.Join(root, filepath.Base(file))
		if err := RenderFromManifestO(file, into, sm); err != nil {
			return err
		}
	}

	return nil
}

// Renders a template using values from the
// `ServiceManifest` returning bytes of the rendered
// template.
func RenderFromManifest(tmpl *template.Template, sm *manifest.ServiceManifest) ([]byte, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, sm); err != nil {
		return []byte{}, err
	}
	return buf.Bytes(), nil
}

// Loads a template from some file path and then renders the
// template into a slice of bytes using the
// `ServiceManifest`.
func RenderFromManifestF(templateName string, sm *manifest.ServiceManifest) ([]byte, error) {
	tmpl, err := LoadTemplate(templateName)
	if err != nil {
		return []byte{}, err
	}
	return RenderFromManifest(tmpl, sm)
}

// Loads a template from some file path and renders the
// output into another file using values from the
// `ServiceManifest`.
func RenderFromManifestO(from, into string, sm *manifest.ServiceManifest) error {
	render, err := RenderFromManifestF(from, sm)
	if err != nil {
		return err
	}
	return os.WriteFile(into, render, defaultFileMode)
}
