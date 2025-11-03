package service

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/WilkinsonK/grawp/grawpadmin/util"
)

type Archiver struct {
	MaxDepth uint
	Options  ArchiveOpts
}

func (A *Archiver) Archive() error {
	A.Options.InitTar()
	err := A.AddDirectory(A.Options.RootPath)
	if err != nil {
		return err
	}
	A.Options.Close()
	return A.DumpArchive()
}

func (A *Archiver) AddDirectory(path string) error {
	return filepath.WalkDir(path, A.AddDirectoryWalker)
}

func (A *Archiver) AddDirectoryWalker(path string, d fs.DirEntry, err error) error {
	if err != nil || !A.PathIsAllowed(path) || d.Type().IsDir() {
		return err
	}
	return A.AddFile(path)
}

func (A *Archiver) AddExcludes(patterns ...string) {
	A.AddGlobPatterns(&A.Options.Exclude, patterns...)
}

func (A *Archiver) AddIncludes(patterns ...string) {
	A.AddGlobPatterns(&A.Options.Include, patterns...)
}

func (A *Archiver) AddGlobPattern(container *[]string, pattern string) {
	var parts []string

	depth := int(A.MaxDepth) - len(strings.Split(pattern, "**"))
	for ; depth > 0; depth-- {
		part := strings.ReplaceAll(pattern, "**", strings.Repeat("/*", int(depth)))
		part = strings.ReplaceAll(part, "//", "/")
		part = filepath.Join(A.Options.RootPath, part)
		parts = append(parts, part)
	}

	*container = append(*container, parts...)
}

func (A *Archiver) AddGlobPatterns(container *[]string, patterns ...string) {
	util.ForEach(slices.Values(patterns), func(p string) {
		A.AddGlobPattern(container, p)
	})
}

func (A *Archiver) AddFile(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(fi, "")
	header.Name = path
	if err != nil {
		return err
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err = A.Options.Tar.WriteHeader(header); err != nil {
		return err
	}
	if _, err = A.Options.Tar.Write(contents); err != nil {
		return err
	}
	fmt.Println("Added: ", path)
	return nil
}

func (A *Archiver) GetMaxDepth(path string) (uint, error) {
	return A.MaxDepth, filepath.WalkDir(path, A.GetMaxDepthWalker)
}

func (A *Archiver) GetMaxDepthWalker(path string, d fs.DirEntry, err error) error {
	if err != nil || d.Type().IsDir() {
		return err
	}
	A.MaxDepth = max(A.MaxDepth, uint(len(strings.Split(path, "/"))))
	return nil
}

func (A *Archiver) Close() error {
	A.Options.Buffer.Reset()
	return A.Options.Close()
}

func (A *Archiver) DumpArchive() error {
	opts := A.Options
	name := filepath.Join(opts.Path, opts.Name)
	byts := opts.Buffer.Bytes()

	if len(byts) <= 32 {
		return fmt.Errorf("No files to archive a %s", opts.RootPath)
	}

	err := os.WriteFile(name, byts, 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to file %s: %s\n", name, err)
		return err
	}
	fmt.Printf("Archive written to %s\n", name)
	return nil
}

func (A *Archiver) PathIsAllowed(path string) bool {
	return !A.PathIsExcluded(path) && A.PathIsIncluded(path)
}

func (A *Archiver) PathIsExcluded(path string) bool {
	if len(A.Options.Exclude) == 0 {
		return false
	}

	var notAllowed bool = false
	for _, pattern := range A.Options.Exclude {
		if yes, err := filepath.Match(pattern, path); err != nil || !yes {
			continue
		}
		notAllowed = true
	}

	return notAllowed
}

func (A *Archiver) PathIsIncluded(path string) bool {
	if len(A.Options.Include) == 0 {
		return true
	}

	var isAllowed bool = false
	for _, pattern := range A.Options.Include {
		if yes, err := filepath.Match(pattern, path); err != nil || !yes {
			continue
		}
		isAllowed = true
	}

	return isAllowed
}

type ArchiveOpts struct {
	Buffer bytes.Buffer
	// Patterns to include in archive.
	Include []string
	// Patterns to exclude from archive.
	Exclude  []string
	Name     string
	Path     string
	RootPath string
	Gzip     *gzip.Writer
	Tar      *tar.Writer
}

func (Ao *ArchiveOpts) Close() error {
	Ao.Tar.Close()
	Ao.Gzip.Close()
	return nil
}

func (Ao *ArchiveOpts) InitTar() {
	Ao.Gzip = gzip.NewWriter(&Ao.Buffer)
	Ao.Tar = tar.NewWriter(Ao.Gzip)
}

func ArchiverNew(opts ArchiveOpts) Archiver {
	a := Archiver{Options: opts}
	a.GetMaxDepth(a.Options.RootPath)
	return a
}

func ArchiveOptsNew(name, path, rootPath string) ArchiveOpts {
	ao := ArchiveOpts{
		Name:     name,
		Path:     path,
		RootPath: rootPath,
	}
	ao.InitTar()
	return ao
}
