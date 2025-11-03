package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

var DeveloperMode string
var ProjectRootPath string

var _DeveloperMode bool
var _ProjectRootPath string

func initLinkOptions() {
	yes, err := strconv.ParseBool(DeveloperMode)
	if err != nil {
		_DeveloperMode = false
	}
	_DeveloperMode = yes

	if !_DeveloperMode {
		DeveloperMode = ""
		ProjectRootPath = ""
		return
	}

	_ProjectRootPath = ProjectRootPath
}

func DoRebuildSelf() error {
	os.Chdir(_ProjectRootPath)
	cmd := exec.Command("make")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n%s", err, string(output))
		return err
	}
	fmt.Print(string(output))
	return nil
}
