package main

import (
	"github.com/alberttduong/card-game/game"
	"github.com/bit101/go-ansi"
	"fmt"
	"bufio"
	"os"
	"log"
	"strings"
	"encoding/json"
	_ "embed"
)

//go:embed card_data.json 

var data []byte

func GetCardData() map[string]game.Cdata {
	var rawCards map[string]json.RawMessage
	err := json.Unmarshal(data, &rawCards) 
	if err != nil {
		log.Fatal(err)
	}

	var cards map[string]game.Cdata = make(map[string]game.Cdata)
	for key := range rawCards {
		var c game.Cdata
		err := json.Unmarshal(rawCards[key], &c)
		if err != nil {
			log.Fatal(err)
		}

		cards[key] = c
	}
	return cards	
}

func main() {
	g, err := game.InitState(2)
	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(os.Stdin)

	ansi.ClearScreen()
	fmt.Println(g)
	fmt.Println()

	var out []string
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		input = strings.Trim(input, "\n")
		if input == "q" {
			ansi.ClearScreen()
			break
		}

		args := strings.Split(input, " ")
		
		g, out = game.Execute(g, args...)

		ansi.ClearScreen()
		fmt.Println(g)
		fmt.Println()
		fmt.Println(out)
		fmt.Println()
	}
}
