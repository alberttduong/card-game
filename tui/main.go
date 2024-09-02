package main

import (
	"github.com/alberttduong/card-game/game"
	"fmt"
	"bufio"
	"os"
	"log"
	"strings"
	_ "embed"
)

func clearScreen() {
	fmt.Print("\033[H\r")
	fmt.Print("\033[2J\r")	
}

func ExecuteCommands (scan *bufio.Scanner, cards []game.Cdata, g game.State) (newG game.State, e error) {
	scan.Scan() 
	input := scan.Text()

	input = strings.Trim(input, "\n")

	args := strings.Split(input, " ")

	newG, e = g.Execute(cards, args...)
	return
}

//go:embed cards.json 
var data []byte

func main() {
	cards := game.GetCardData(data) 	

	g, err := game.NewTestGame(2)
	if err != nil {
		log.Fatal(err)
	}

	input := bufio.NewScanner(os.Stdin)

	clearScreen()
	fmt.Println(g)
	fmt.Println()	

	var e error
	for {
		if input.Text() == "q" {
			clearScreen()
			break
		}
		g, e = ExecuteCommands(input, cards, g)
		clearScreen()
		fmt.Println(g)
		fmt.Println()
		fmt.Println()
		if e != nil {
			fmt.Printf("Error: %v\n", e)
		}
	}
}
