package main

import (
	"github.com/alberttduong/card-game/game"
	"github.com/alberttduong/card-game/tui"
	"github.com/nsf/termbox-go"
	"log"
	_ "embed"
)

//go:embed cards.json 
var data []byte

func main() {
	err := termbox.Init()
    if err != nil {
        panic(err)
    }
    defer termbox.Close()


	g, err := game.NewTestGame(2)
	if err != nil {
		log.Fatal(err)
	}

	cards := game.GetCardData(data) 	

	//screen := tui.NewDeckBuilder(cards)
	screen := tui.InitScreen(cards)
	screen.Game = tui.NewScreen(cards, g)

	screen.Redraw()

MainLoop:
	for { 
		switch ev := termbox.PollEvent(); ev.Type {
        case termbox.EventKey:
			err = screen.HandleEvent(ev)
			if err == tui.EXIT {
				break MainLoop
			}	
		}
	}
}
