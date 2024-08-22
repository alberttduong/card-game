package game

import (
	"strconv"
)

func Execute(g State, args ...string) (State, []string) {
	out := make([]string, 0)

	sayOut := func(l string) {
		out = append(out, l)
	}

	switch args[0] {
	case "play":
		h := g.players[g.currentPlayer].hand

		if len(args) < 2 {
			sayOut("Not enough args, expected 2")
			return g, out
		}

		idx, err := strconv.Atoi(args[1])
		if err != nil {
			sayOut(err.Error())
			return g, out
		}
		if idx < 0 || idx > len(h)-1 {
			sayOut("Index out of bounds") 
			return g, out
		}

		//c := h[idx]
		card := Card{0, 3, 4}
		game, err := g.play(g.currentPlayer, card)
		if err != nil {
			sayOut(err.Error())
			return g, out
		}

		game, err = game.removeFromHand(game.currentPlayer, idx)
		if err != nil {
			sayOut(err.Error())
			return g, out
		}

		return game, out 
	case "draw":
		g, err := g.drawCard(g.currentPlayer)
		if err != nil {
			sayOut(err.Error())
			return g, out
		}
		sayOut("Drew 1 card")
		return g, out
	case "end":
		g, err := g.endTurn()
		if err != nil {
			sayOut(err.Error())
			return g, out
		}
		return g, out
	default:
		sayOut("Invalid command") 
	}
	return g, out 
}
