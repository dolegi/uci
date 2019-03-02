package uci

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type engine struct {
	stdin  *bufio.Writer
	stdout *bufio.Scanner
	moves  string
	Meta   Meta
	Side   int
	Type   int
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

type NewGameOpts struct {
	Type int
	Side int
}

type GoOpts struct {
	SearchMoves string
	Ponder      bool
	Wtime       int
	Btime       int
	Winc        int
	Binc        int
	MovesToGo   int
	Depth       int
	Nodes       int
	Mate        int
	MoveTime    int
}

type GoResp struct {
	Bestmove string
	Ponder   string
}

const (
	ALG int = 0
	FEN int = 1
	W   int = 0
	B   int = 1
)

var execCommand = exec.Command

func NewEngine(path string) (*engine, error) {
	eng := engine{}
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
	eng.Meta = eng.getMeta()
	return &eng, nil
}

func (eng *engine) getMeta() (meta Meta) {
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
			meta.Options = append(meta.Options, newOption(strings.TrimPrefix(line, optionPrefix)))
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

func newOption(line string) (option Option) {
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

func (eng *engine) SetOption(name string, value interface{}) bool {
	for _, option := range eng.Meta.Options {
		if option.Name == name {
			var v string
			switch value.(type) {
			case string:
				v, _ = value.(string)
			case int:
				vv, _ := value.(int)
				v = strconv.Itoa(vv)
			case bool:
				vv, _ := value.(bool)
				if vv {
					v = "true"
				} else {
					v = "false"
				}
			}
			eng.send("setoption name " + name + " value " + v)
			return true
		}
	}
	return false
}

func (eng *engine) IsReady() bool {
	eng.send("isready")
	lines := eng.receive("readyok")
	return lines[0] == "readyok"
}

func (eng *engine) send(input string) {
	_, err := eng.stdin.WriteString(input + "\n")
	if err == nil {
		eng.stdin.Flush()
	}
}

func (eng *engine) receive(stopPrefix string) (lines []string) {
	scanner := eng.stdout
	for scanner.Scan() {
		line := scanner.Text()
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

func (eng *engine) NewGame(opts NewGameOpts) {
	if opts.Type == FEN {
		if opts.Side == W {
			eng.send("position fen rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
		} else {
			eng.send("position fen rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1")
		}
		eng.moves = ""
	} else {
		eng.moves = "startpos moves"
		eng.send("position " + eng.moves)
	}
	eng.Type = opts.Type
	eng.Side = opts.Side
}

func (eng *engine) Position(pos string) {
	if eng.Type == FEN {
		eng.send("position fen " + pos)
	} else {
		eng.moves = eng.moves + " " + pos
		eng.send("position " + eng.moves)
	}
}

func addOpt(name string, value int) string {
	if value > 0 {
		return name + " " + strconv.Itoa(value)
	}
	return ""
}

func (eng *engine) Go(opts GoOpts) GoResp {
	goCmd := "go "
	if opts.Ponder {
		goCmd += "ponder "
	}
	goCmd += addOpt("wtime", opts.Wtime)
	goCmd += addOpt("btime", opts.Btime)
	goCmd += addOpt("winc", opts.Winc)
	goCmd += addOpt("binc", opts.Binc)
	goCmd += addOpt("movestogo", opts.MovesToGo)
	goCmd += addOpt("depth", opts.Depth)
	goCmd += addOpt("nodes", opts.Nodes)
	goCmd += addOpt("mate", opts.Mate)
	goCmd += addOpt("movetime", opts.MoveTime)

	eng.send(goCmd)
	lines := eng.receive("bestmove")
	words := strings.Split(lines[len(lines)-1], " ")

	return GoResp{
		Bestmove: words[1],
		Ponder:   words[3],
	}
}

func (eng *engine) Quit() {
	eng.send("quit")
	eng.stdin = nil
	eng.stdout = nil
	eng.Meta = Meta{}
	eng.moves = ""
	eng.Side = 0
	eng.Type = 0
}
