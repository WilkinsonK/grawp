package main

import (
	"fmt"
	"os"
	"os/exec"
)

var ProjectRootPath string

func DoRebuildSelf() error {
	fmt.Println("Rebuilding...")
	os.Chdir(ProjectRootPath)
	fmt.Printf("ProjectRootPath: %s\n", ProjectRootPath)
	cmd := exec.Command("make")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n%s", err, string(output))
		return err
	}
	fmt.Print(string(output))
	return nil
}
