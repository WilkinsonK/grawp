package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/WilkinsonK/grawp/grawpadmin/manifest"
)

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

func RenderFromManifest(tmpl *template.Template, sm *manifest.ServiceManifest) ([]byte, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, sm); err != nil {
		return []byte{}, err
	}
	return buf.Bytes(), nil
}

func RenderFromManifestF(templateName string, sm *manifest.ServiceManifest) ([]byte, error) {
	tmpl, err := LoadTemplate(templateName)
	if err != nil {
		return []byte{}, err
	}
	return RenderFromManifest(tmpl, sm)
}

func RenderFromManifestO(from, into string, sm *manifest.ServiceManifest) error {
	render, err := RenderFromManifestF(from, sm)
	if err != nil {
		return err
	}
	return os.WriteFile(into, render, 0755)
}
