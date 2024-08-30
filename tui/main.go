package main

import (
	"github.com/alberttduong/card-game/game"
	"fmt"
	"bufio"
	"os"
	"log"
	"strings"
	"encoding/json"
	_ "embed"
)

func GetCardData(data []byte) []game.Cdata {
	//var rawCards map[string]json.RawMessage
	var rawCards []json.RawMessage
	err := json.Unmarshal(data, &rawCards) 
	if err != nil {
		log.Fatal(err)
	}

	cards := make([]game.Cdata, 30) 
	for i, v := range rawCards {
		var c game.Cdata
		err := json.Unmarshal(v, &c)
		if err != nil {
			log.Fatal(err)
		}

		cards[i] = c
	}
	return cards	
}

func clearScreen() {
	fmt.Print("\033[H\r")
	fmt.Print("\033[2J\r")	
}

func ExecuteCommands (scan *bufio.Scanner, cards []game.Cdata, g game.State) (newG game.State, newOut []string) {
	scan.Scan() 
	input := scan.Text()

	input = strings.Trim(input, "\n")

	args := strings.Split(input, " ")

	newG, newOut = g.Execute(cards, args...)
	return
}

//go:embed cards.json 
var data []byte

func main() {
	cards := GetCardData(data) 	

	g, err := game.InitState(2)
	if err != nil {
		log.Fatal(err)
	}

	input := bufio.NewScanner(os.Stdin)

	clearScreen()
	fmt.Println(g)
	fmt.Println()	

	var out []string
	for {
		if input.Text() == "q" {
			clearScreen()
			break
		}
		g, out = ExecuteCommands(input, cards, g) 
		clearScreen()
		fmt.Println(g)
		fmt.Println()
		fmt.Println(out)
		fmt.Println()
	}
}
