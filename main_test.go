package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"testing"
	"time"
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
	fmt.Println("hit")
	input, _ := io.Pipe()

	message := []byte{}
	time.Sleep(1500)
	_, err := io.ReadFull(input, message)
	fmt.Println(string(message))
	if err != nil {
		panic(err)
	}
	os.Exit(0)
}
