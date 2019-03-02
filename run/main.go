package main

import (
	"fmt"

	"github.com/dolegi/uci"
	"os"
)

func main() {
	eng, _ := uci.NewEngine(os.Args[1])
	eng.SetOption("Ponder", false)
	eng.SetOption("Threads", "2")
	if eng.IsReady() {
		eng.NewGame(uci.NewGameOpts{})
		eng.Position("e2e4")
		resp := eng.Go(uci.GoOpts{MoveTime: 100})
		fmt.Println(resp.Bestmove)
	}
	eng.Quit()
}
