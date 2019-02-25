package main

import (
	"bufio"
	"os/exec"
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

type GoResponse struct {
}

var eng Engine

// eng = *NewEngine("./stockfish")

var execCommand = exec.Command

func NewEngine(path string) *Engine {
	eng := Engine{}
	cmd := execCommand(path)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil
	}
	if err := cmd.Start(); err != nil {
		return nil
	}
	eng.stdin = bufio.NewWriter(stdin)
	eng.stdout = bufio.NewScanner(stdout)
	eng.send("uci")
	return &eng
}

func (eng *Engine) send(input string) {
	_, err := eng.stdin.WriteString(input + "\n")
	if err == nil {
		eng.stdin.Flush()
	}
}

// func (eng *Engine) NewGame() bool {
// 	eng.send("ucinewgame")
// 	eng.Position("startpos")
// }

// func (eng *Engine) Position(pos string) bool {
// 	eng.moves = eng.moves + " " + pos
// 	eng.send("position" + eng.moves)
// }

// func (eng *Engine) Go(options string) GoResponse {
// }
