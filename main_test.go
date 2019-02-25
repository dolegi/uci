package main

import (
	"os"
	"os/exec"
	"testing"
)

func TestNewEngine(t *testing.T) {
	execCommand = mockExecCmd
	defer func() { execCommand = exec.Command }()

	NewEngine("./path/to/file")
}

func mockExecCmd(cmd string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperCmd", "--", cmd}
	return exec.Command(os.Args[0], cs...)
}

func TestHelperCmd(t *testing.T) {
	os.Exit(0)
}
