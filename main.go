package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// TODO
// get id name
// get options
// new game
// set option
// check if ready
// position
// go
// stop
// ponderhit
// quit

type Engine struct {
	stdin  *bufio.Writer
	stdout *bufio.Scanner
	moves  string
}

type Meta struct {
	Name    string
	Author  string
	Options []Option
}

type Option struct {
	Name    string
	Type    string
	Default interface{}
	Min     int
	Max     int
}

var execCommand = exec.Command

func NewEngine(path string) (*Engine, error) {
	eng := Engine{}
	cmd := execCommand(path)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	eng.stdin = bufio.NewWriter(stdin)
	eng.stdout = bufio.NewScanner(stdout)
	return &eng, nil
}

func (eng *Engine) GetMeta() (meta Meta) {
	lines := eng.send("uci", "uciok")

	namePrefix := "id name "
	authorPrefix := "id author "
	optionPrefix := "option "
	for _, line := range lines {
		if strings.HasPrefix(line, namePrefix) {
			meta.Name = strings.TrimPrefix(line, namePrefix)
		} else if strings.HasPrefix(line, authorPrefix) {
			meta.Author = strings.TrimPrefix(line, authorPrefix)
		} else if strings.HasPrefix(line, optionPrefix) {
			meta.Options = append(meta.Options, NewOption(strings.TrimPrefix(line, optionPrefix)))
		}
	}
	return meta
}

func NewOption(line string) (option Option) {
	nameRegex := regexp.MustCompile(`name (.*) type`)
	typeRegex := regexp.MustCompile(`type (\w+)`)

	option.Name = nameRegex.FindStringSubmatch(line)[1]
	option.Type = typeRegex.FindStringSubmatch(line)[1]

	return
}

func (eng *Engine) send(input, stopPrefix string) (lines []string) {
	_, err := eng.stdin.WriteString(input + "\n")
	if err == nil {
		eng.stdin.Flush()
	}

	return eng.receive(stopPrefix)
}

func (eng *Engine) receive(stopPrefix string) (lines []string) {
	for eng.stdout.Scan() {
		line := eng.stdout.Text()
		lines = append(lines, line)
		if strings.HasPrefix(line, stopPrefix) {
			break
		}
	}
	if err := eng.stdout.Err(); err != nil {
		fmt.Println("reading standard input:", err)
	}
	return
}

func main() {
	eng, _ := NewEngine("./stockfish")
	meta := eng.GetMeta()
	fmt.Println(meta)
}
