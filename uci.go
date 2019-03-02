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

// Meta data about the engine
type Meta struct {
	Name    string   // Name of engine
	Author  string   // Author of the engine
	Options []Option // Available options for this engine
}

// Available options to set on this engine
type Option struct {
	Name    string      // Name of option
	Type    string      // Type of option
	Default interface{} // Default value of option
	Min     int         // Min value of option
	Max     int         // Max value of option
	Vars    []string    // Enum vars for option
}

// Options for creating a new game
type NewGameOpts struct {
	Type int // Type of positioning. Must be uci.FEN or uci.ALG
	Side int // Which side should the engine play as. Must be uci.W or uci.B
}

// Options to pass when looking for best move
type GoOpts struct {
	SearchMoves string // <move1> .... <movei>. restrict search to this moves only
	Ponder      bool   // start searching in pondering mode
	Wtime       int    // number of ms white has left
	Btime       int    // number of ms black has left
	Winc        int    // number of ms white increases by each move
	Binc        int    // number of ms black increases by each move
	MovesToGo   int    // number of moves until next time control
	Depth       int    // maximum search depth
	Nodes       int    // maximum search nodes
	Mate        int    // search for mate in x moves
	MoveTime    int    // number of ms exactly to search for
}

// Response from searching
type GoResp struct {
	Bestmove string // Best move the engine could find
	Ponder   string // Ponder move
}

const (
	ALG int = 0 // Type of positioning. Full algorithmic positioning e.g e2e4 d7d6 ...
	FEN int = 1 // Type of positioning. FEN string
	W   int = 0 // Side to play as. White
	B   int = 1 // Side to play as. Black
)

var execCommand = exec.Command

// Create a new engine requires the path to a uci chess engine such as stockfish
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

// Pass an option to the underlying engine
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

// Check if engine is ready to start receiving commands
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

// Start a new game. Only one game should be played at a time
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

// Set the position of the game. Either full fen string or next position such as "e2e4"
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

// Search for the bestmove
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

// Quit the engine. Engine struct cannot be used after this command has been sent
func (eng *engine) Quit() {
	eng.send("quit")
	eng.stdin = nil
	eng.stdout = nil
	eng.Meta = Meta{}
	eng.moves = ""
	eng.Side = 0
	eng.Type = 0
}
