package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// TODO
// add option vars
// test isready
// new game
// position
// go
// stop
// ponderhit
// quit

type Engine struct {
	stdin  *bufio.Writer
	stdout *bufio.Scanner
	moves  string
	Meta   Meta
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
	Vars    []string
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
	eng.Meta = eng.GetMeta()
	return &eng, nil
}

func (eng *Engine) GetMeta() (meta Meta) {
	eng.send("uci")
	lines := eng.receive("uciok")

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

func getOption(line, regex string) interface{} {
	rr := regexp.MustCompile(regex)
	results := rr.FindStringSubmatch(line)

	if len(results) == 2 {
		return results[1]
	}
	return nil
}

func NewOption(line string) (option Option) {
	option.Name, _ = getOption(line, `name (.*) type`).(string)
	option.Type, _ = getOption(line, `type (\w+)`).(string)
	option.Default = getOption(line, `default (\w+)`)
	minStr, _ := getOption(line, `min (\w+)`).(string)
	maxStr, _ := getOption(line, `max (\w+)`).(string)
	option.Min, _ = strconv.Atoi(minStr)
	option.Max, _ = strconv.Atoi(maxStr)

	varRegex := regexp.MustCompile(`var (\w+)`)

	vars := []string{}
	for _, v := range varRegex.FindAllStringSubmatch(line, -1) {
		vars = append(vars, v[1])
	}
	option.Vars = vars

	return
}

func (eng *Engine) SetOption(name, value string) bool {
	for _, option := range eng.Meta.Options {
		if option.Name == name {
			eng.send("setoption name " + name + " value " + value)
			return true
		}
	}
	return false
}

func (eng *Engine) IsReady() bool {
	eng.send("isready")
	lines := eng.receive("readyok")
	return lines[0] == "readyok"
}

func (eng *Engine) send(input string) {
	_, err := eng.stdin.WriteString(input + "\n")
	if err == nil {
		eng.stdin.Flush()
	}
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
	passed := eng.SetOption("Threads", "10")
	fmt.Println(passed)

	fmt.Println(eng.IsReady())
}
