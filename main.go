package main

import (
	"github.com/DominicGinger/uci"
)

func main() {
	eng, _ := uci.NewEngine("./stockfish")
	eng.SetOption("Ponder", false)
	eng.SetOption("Threads", "2")
	if eng.IsReady() {
		eng.NewGame(NewGameOpts{})
		eng.Position("e2e4")
		resp := eng.Go(GoOpts{MoveTime: 100})
		fmt.Println(resp.Bestmove)
	}
	eng.Quit()
}
