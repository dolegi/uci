package main

import (
	"log"
	"os"
	"os/exec"
	"reflect"
	"testing"
)

func mockExecCmd(cmd string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperCmd", "--", cmd}
	c := exec.Command(os.Args[0], cs...)
	c.Env = []string{"EXEC_TEXT=1"}
	return c
}

func TestHelperCmd(t *testing.T) {
	if os.Getenv("EXEC_TEST") != "1" {
		return
	}
	os.Exit(0)
}

func TestNewEngine(t *testing.T) {
	execCommand = mockExecCmd
	defer func() { execCommand = exec.Command }()

	eng, err := NewEngine("./path/to/file")

	if err != nil {
		log.Fatal("NewEngine Returned error", err)
	}

	if eng.Meta.Name != "" ||
		eng.Meta.Author != "" ||
		len(eng.Meta.Options) != 0 {
		log.Fatal("NewEngine New meta isn't empty")
	}
}

func TestNewEngineStockfish(t *testing.T) {
	eng, err := NewEngine("./stockfish")
	expectedMeta := Meta{
		Name:   "Stockfish 160219 64 POPCNT",
		Author: "T. Romstad, M. Costalba, J. Kiiski, G. Linscott",
		Options: []Option{
			{Name: "Debug Log File", Type: "string", Vars: []string{}},
			{Name: "Contempt", Type: "spin", Default: "24", Min: 0, Max: 100, Vars: []string{}},
			{Name: "Analysis Contempt", Type: "combo", Default: "Both", Vars: []string{"Off", "White", "Black", "Both"}},
			{Name: "Threads", Type: "spin", Default: "1", Min: 1, Max: 512, Vars: []string{}},
			{Name: "Hash", Type: "spin", Default: "16", Min: 1, Max: 131072, Vars: []string{}},
			{Name: "Clear Hash", Type: "button", Vars: []string{}},
			{Name: "Ponder", Type: "check", Default: "false", Vars: []string{}},
			{Name: "MultiPV", Type: "spin", Default: "1", Min: 1, Max: 500, Vars: []string{}},
			{Name: "Skill Level", Type: "spin", Default: "20", Min: 0, Max: 20, Vars: []string{}},
			{Name: "Move Overhead", Type: "spin", Default: "30", Min: 0, Max: 5000, Vars: []string{}},
			{Name: "Minimum Thinking Time", Type: "spin", Default: "20", Min: 0, Max: 5000, Vars: []string{}},
			{Name: "Slow Mover", Type: "spin", Default: "84", Min: 10, Max: 1000, Vars: []string{}},
			{Name: "nodestime", Type: "spin", Default: "0", Min: 0, Max: 10000, Vars: []string{}},
			{Name: "UCI_Chess960", Type: "check", Default: "false", Vars: []string{}},
			{Name: "UCI_AnalyseMode", Type: "check", Default: "false", Vars: []string{}},
			{Name: "SyzygyPath", Type: "string", Vars: []string{}},
			{Name: "SyzygyProbeDepth", Type: "spin", Default: "1", Min: 1, Max: 100, Vars: []string{}},
			{Name: "Syzygy50MoveRule", Type: "check", Default: "true", Vars: []string{}},
			{Name: "SyzygyProbeLimit", Type: "spin", Default: "7", Min: 0, Max: 7, Vars: []string{}},
		},
	}

	if err != nil {
		log.Fatal("NewEngineStockfish Returned error", err)
	}

	if !reflect.DeepEqual(eng.Meta, expectedMeta) {
		log.Fatal("NewEngineStockfish meta does not match", eng.Meta, expectedMeta)
	}
}

func TestIsReady(t *testing.T) {
	eng, _ := NewEngine("./stockfish")
	if eng.IsReady() != true {
		log.Fatal("TestIsReady did not return true")
	}
}

func TestNewGame(t *testing.T) {
	eng, _ := NewEngine("./stockfish")
	eng.NewGame(NewGameOpts{})

	if eng.moves != "" {
		log.Fatal("TestNewGame too many moves")
	}
}

func TestPositionFEN(t *testing.T) {
	eng, _ := NewEngine("./stockfish")
	eng.NewGame(NewGameOpts{})
	eng.Position("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")

	if eng.moves != "" {
		log.Fatal("TestPositionFEN too many moves")
	}
}

func TestPosition(t *testing.T) {
	eng, _ := NewEngine("./stockfish")
	eng.NewGame(NewGameOpts{1, 0})
	eng.Position("e2e4")
	eng.Position("d7d6")

	if eng.moves != " e2e4 d7d6" {
		log.Fatal("TestPosition wrong amount of moves")
	}
}
