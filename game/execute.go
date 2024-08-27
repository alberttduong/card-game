package game

import (
	"strconv"
	"errors"
)

type argType int
const (
	Command argType = iota
	CardId
	Int
	CurrFieldId
	PlayerId
	FieldId
	AtkId
)

func verifyArgs(cards []Cdata, expected ...argType) func (...string) error {
	return func (args ...string) error {
		if len(args) < len(expected) {
			return errors.New("Unexpected number of arguments") 
		}

		for i, t := range expected {
			if t == Command {
				continue
			}

			n, err := strconv.Atoi(args[i])
			if err != nil {
				return errors.New("Couldn't parse arg") 
			}

			switch t {
			case CardId:
				if !IsValidCardId(n, cards) {
					return errors.New("Invalid CardId")
				}
			}
		}

		return nil
	}
}

func (g State) Execute(cards []Cdata, args ...string) (State, []string) {
	out := make([]string, 0)

	sayOut := func(l string) {
		out = append(out, l)
	}

	switch args[0] {
	case "create":
		// create [CardName]
		err := verifyArgs(cards, Command)(args...) 
		if err != nil {
			sayOut(err.Error())
			return g, out
		}

		n, _ := strconv.Atoi(args[1])

		game, err := g.play(g.currentPlayer, CardFromName(cards, CardName(n)))
		if err != nil {
			sayOut(err.Error())
			return g, out
		}

		return game, out 
	case "setmana":
		// setmana [Int]
		err := verifyArgs(cards, Command, Int)(args...) 
		if err != nil {
			sayOut(err.Error())
			return g, out
		}
		
		n, _ := strconv.Atoi(args[1])
		return g.setMana(n), out
	case "attack":
		// attack [CurrFieldId] [PlayerId] [FieldId] [AtkId]
		err := verifyArgs(cards, Command, CurrFieldId, PlayerId, FieldId, AtkId)(args...)
		if err != nil {
			sayOut(err.Error())
			return g, out
		}

		return g, out
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

		card := CardFromName(cards, h[idx])
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
