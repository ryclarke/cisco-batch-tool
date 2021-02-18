// +build mage

package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Default target to run when none is specified
// If not set, running mage will list available targets
// var Default = Build

var Aliases = map[string]interface{}{
	"build":   Build.All,
	"package": Package.All,
}

var targets = map[string]string{
	"linux":   "bin/linux/batch-tool",
	"darwin":  "bin/darwin/batch-tool",
	"windows": "bin/windows/batch-tool.exe",
}

var releases = map[string]string{
	"linux":   "release/batch-tool-linux.tgz",
	"darwin":  "release/batch-tool-darwin.tgz",
	"windows": "release/batch-tool-windows.zip",
}

type Build mg.Namespace

// Build the executables for all three supported platforms
func (Build) All() {
	// Declares target dependencies that will be run in parallel
	mg.Deps(Build.Linux, Build.Darwin, Build.Windows)
}

// Build the executable for Linux
func (Build) Linux() error {
	return goBuild("linux")
}

// Build the executable for MacOS
func (Build) Darwin() error {
	return goBuild("darwin")
}

// Build the executable for Windows
func (Build) Windows() error {
	return goBuild("windows")
}

type Package mg.Namespace

// Package the executables for all three supported platforms
func (Package) All() {
	mg.Deps(Package.Linux, Package.Darwin, Package.Windows)
}

// Package the executable for Linux
func (Package) Linux() error {
	mg.Deps(Build.Linux)

	return archiveCmd("linux")
}

// Package the executable for MacOS
func (Package) Darwin() error {
	mg.Deps(Build.Darwin)

	return archiveCmd("darwin")
}

// Package the executable for Windows
func (Package) Windows() error {
	mg.Deps(Build.Windows)

	return archiveCmd("windows")
}

// Install the executable to the local environment
func Install() error {
	// determine the installation path based on environment variables
	installDir := os.Getenv("GOBIN")
	if installDir == "" {
		if goPath := os.Getenv("GOPATH"); goPath != "" {
			installDir = filepath.Join(goPath, "bin")
		} else {
			homedir, err := os.UserHomeDir()
			if err != nil {
				return err
			}

			installDir = filepath.Join(homedir, "go", "bin")
		}
	}

	return sh.Run("go", "build", "-i", "-o", filepath.Join(installDir, "batch-tool"))
}

// Clean up after yourself
func Clean() {
	sh.Rm("bin")
	sh.Rm("release")
}

// helper function to populate build sources with golang source files
func goSources() []string {
	sources := make([]string, 0)

	// use go list to get list of files with and without mage build target, then calculate the difference
	// to exclude mage files.
	nomage, err := sh.Output("go", "list", "-e", "-f={{join .GoFiles \"\\n\"}}", "./...")
	if err != nil {
		fmt.Printf("error fetching go source files: %v\n", err)

		return []string{}
	}

	nomageslice := strings.Fields(nomage)

	mage, err := sh.Output("go", "list", "-tags=mage", "-e", "-f={{join .GoFiles \"\\n\"}}", "./...")
	if err != nil {
		fmt.Printf("error fetching go source files: %v\n", err)

		return []string{}
	}

	mageslice := strings.Fields(mage)

	// here we are constructing a set using a map, where values are empty structs
	mageset := make(map[string]struct{})
	for _, elem := range mageslice {
		mageset[elem] = struct{}{}
	}

	// remove any keys that match the nomage list.
	for _, elem := range nomageslice {
		delete(mageset, elem)
	}

	cwd, _ := os.Getwd()
	if err := filepath.Walk(cwd, func(path string, f os.FileInfo, err error) error {
		if err != nil || f.IsDir() {
			return err
		}

		// match all golang source files (except for the magefiles)
		isGoFile, _ := filepath.Match("*.go", f.Name())
		_, isMageFile := mageset[f.Name()]
		if isGoFile && !isMageFile {
			sources = append(sources, path)
		}

		return nil
	}); err != nil {
		fmt.Printf("error fetching go source files: %v\n", err)

		return []string{}
	}

	return sources
}

// helper function to call go build with the correct os and target
func goBuild(targetOS string) error {
	if updated, err := target.Path(targets[targetOS], goSources()...); !updated || err != nil {
		return err
	}

	env := map[string]string{
		"GOARCH": "amd64",
		"GOOS":   targetOS,
	}

	return sh.RunWith(env, "go", "build", "-o", targets[targetOS])
}

// Helper function for creating release archives
func archiveCmd(targetOS string) error {
	if updated, err := target.Path(releases[targetOS], targets[targetOS]); !updated || err != nil {
		return err
	}

	// open the executable file for reading
	file, err := os.Open(targets[targetOS])
	if err != nil {
		return err
	}
	defer file.Close()

	// ensure that the archive destination exists
	if err := os.MkdirAll(filepath.Dir(releases[targetOS]), 0755); err != nil {
		return err
	}

	// create (or truncate) the release archive
	archive, err := os.Create(releases[targetOS])
	if err != nil {
		return err
	}
	defer archive.Close()

	// determine archive method based on file extension
	switch filepath.Ext(releases[targetOS]) {
	case ".zip":
		return writeZip(archive, file)
	case ".tar":
		return writeTar(archive, file, false)
	case ".tar.gz", ".tgz":
		return writeTar(archive, file, true)
	}

	return fmt.Errorf("unsupported archive format: %v", filepath.Ext(releases[targetOS]))
}

// Helper function to write file contents to tar, with or without gzip compression
func writeTar(archive *os.File, file *os.File, useGzip bool) error {
	var tw *tar.Writer

	if useGzip {
		// initialize writers for tar and gzip
		gzw := gzip.NewWriter(archive)
		defer gzw.Close()

		tw = tar.NewWriter(gzw)
	} else {
		// initialize tar writer alone
		tw = tar.NewWriter(archive)
	}
	defer tw.Close()

	fi, err := file.Stat()
	if err != nil {
		return err
	}

	// generate tar header based on the provided FileInfo
	header, err := tar.FileInfoHeader(fi, fi.Name())
	if err != nil {
		return err
	}

	// add header to tar writer
	if err = tw.WriteHeader(header); err != nil {
		return err
	}

	// add file contents to tar writer
	_, err = io.Copy(tw, file)
	return err
}

// Helper function to write file contents to zip
func writeZip(archive *os.File, file *os.File) error {
	// initalize writer for tar
	zw := zip.NewWriter(archive)
	defer zw.Close()

	fi, err := file.Stat()
	if err != nil {
		return err
	}

	// generate zip header based on the provided FileInfo
	header, err := zip.FileInfoHeader(fi)
	if err != nil {
		return err
	}

	// add header to zip writer with compression
	header.Method = zip.Deflate
	writer, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}

	// add file contents to zip writer
	_, err = io.Copy(writer, file)
	return err
}
