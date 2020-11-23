// +build mage

package main

import (
	"fmt"
	"github.com/magefile/mage/mg" // mg contains helpful utility functions, like Deps
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
	"os"
	"path/filepath"
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
	"linux":   "release/batch-tool-linux.txz",
	"darwin":  "release/batch-tool-darwin.txz",
	"windows": "release/batch-tool-windows.zip",
}

type Build mg.Namespace

// helper function to call go build with the correct os and target
func goBuild(goOS string) error {
	env := map[string]string{
		"GOARCH": "amd64",
		"GOOS":   goOS,
	}
	// TODO(dakojohn): Should we provide an Install option as well to directly install the program using `go install`?
	return sh.RunWith(env, "go", "build", "-o", targets[goOS])
}

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

// Helper function for the tar command for MacOS and Linux.  Command should work on both platforms.
func tarCmd(targetOS string) error {
	// While much more verbose, this could be fleshed out with a go-native implementation that would be cross-platform compatible.
	// Relevant standard library packages would be archive and compress.
	dir := fmt.Sprintf("-C%s", filepath.Dir(targets[targetOS]))
	file := filepath.Base(targets[targetOS])
	return sh.Run("tar", "-caf", releases[targetOS], dir, file)
}

// Package the executables for all three supported platforms
func (Package) All() {
	mg.Deps(Package.Linux, Package.Darwin, Package.Windows)
}

// Package the executable for Linux
func (Package) Linux() error {
	// Unlike `mg.Deps()` which declares the target functions to run as dependencies, the target package operates like make file-targets.
	// It checks if the sources have been modified more recently than the destination.
	// Params are `target.Path([destination], [sources...])`.
	if updated, err := target.Path(releases["linux"], targets["linux"]); !updated || err != nil {
		return err
	}
	if err := os.MkdirAll("release", 0755); err != nil {
		return err
	}
	return tarCmd("linux")
}

// Package the executable for MacOS
func (Package) Darwin() error {
	if updated, err := target.Path(releases["darwin"], targets["darwin"]); !updated || err != nil {
		return err
	}
	if err := os.MkdirAll("release", 0755); err != nil {
		return err
	}
	return tarCmd("darwin")
}

// Package the executable for Windows
func (Package) Windows() error {
	if updated, err := target.Path(releases["windows"], targets["windows"]); !updated || err != nil {
		return err
	}
	if err := os.MkdirAll("release", 0755); err != nil {
		return err
	}
	return sh.Run("zip", "-j", releases["windows"], targets["windows"])
}

// Clean up after yourself
func Clean() {
	//fmt.Println("Cleaning...")
	sh.Rm("bin")
	sh.Rm("release")
}
