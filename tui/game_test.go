package main 

import (
	"github.com/alberttduong/card-game/game"
	"strings"
	"bufio"
	"testing"
	"fmt"
)

func Test_SetMana(t *testing.T) {
	g := setup(game.Librarian)
	g = execute(g, "setmana 6")
	expected := 6
	if g.Mana != expected {
		t.Errorf("Expected: %d, Got: %d", expected, g.Mana)
	}
}

func Test_Librarian(t *testing.T) {
	g := setup(game.Librarian)
	g = execute(g, "attack 0 0 0 0")
	fmt.Println(g)
}

func Test_Magician(t *testing.T) {
	g := setup(game.Magician)
	g = execute(g, "create 1")
	fmt.Println(g)
}

func execute(g game.State, cmd string) game.State {
	input := bufio.NewScanner(strings.NewReader(cmd))
	g, _ = ExecuteCommands(input, cards(), g)
	return g 
}

func cards() []game.Cdata {
	return GetCardData(data)  
}

func setup(n game.CardName) game.State {
	g, _ := game.InitState(2)

	cmdStr := fmt.Sprintf("create 1\nend\ncreate %d", n) 
	cmds := strings.Split(cmdStr, "\n")

	for _, c := range cmds {
		input := bufio.NewScanner(strings.NewReader(c))
		g, _ = ExecuteCommands(input, cards(), g)
	}

	return g
}
