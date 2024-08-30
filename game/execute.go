package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

type argType int

const (
	Command argType = iota
	Int
	CardId
	CurrFieldId
	PlayerId
	FieldId
	AtkId
)

func convertArgs(n int, args ...string) ([]int, error) {
	lenArgs := len(args)
	if lenArgs < n {
		return []int{}, errors.New(fmt.Sprintf("expected %d args, got %d", n, lenArgs))
	}

	result := make([]int, len(args))
	var err error
	for i, v := range args {
		result[i], err = strconv.Atoi(v)
		if err != nil {
			return []int{}, err
		}
	}
	return result, nil
}

func (g State) Execute(cards []Cdata, args ...string) (State, []string, error) {
	out := make([]string, 0)
	sayOut := func(l string) {
		out = append(out, l)
	}

	switch args[0] {
	case "target":
		nums, err := convertArgs(2, args[1:]...)
		if err != nil {
			return g, out, err
		}

		g, err = g.target(target{pID: playerID(nums[0]), id: nums[1]})
		if err != nil {
			return g, out, err
		}

		return g, out, nil
	case "create":
		nums, err := convertArgs(1, args[1:]...)
		if err != nil {
			return g, out, err
		}

		game, err := g.play(g.currentPlayer, CardFromName(cards, CardName(nums[0])))
		if err != nil {
			return g, out, err
		}

		return game, out, nil
	case "setmana":
		nums, err := convertArgs(1, args[1:]...)
		if err != nil {
			return g, out, err
		}

		return g.setMana(nums[0]), out, nil
	case "attack":
		nums, err := convertArgs(5, args[1:]...)
		if err != nil {
			return g, out, err
		}

		newG, err := g.attack(
			target{
				pID:    playerID(nums[0]),
				id:     nums[1],
				atkNum: nums[2],
			},
			target{
				pID: playerID(nums[3]),
				id:  nums[4],
			})

		if err != nil {
			return g, out, err
		}

		return newG, out, nil
	case "play":
		h := g.players[g.currentPlayer].hand

		if len(args) < 2 {
			sayOut("Not enough args, expected 2")
			return g, out, errors.New("temp")
		}

		idx, err := strconv.Atoi(args[1])
		if err != nil {
			return g, out, err
		}
		if idx < 0 || idx > len(h)-1 {
			return g, out, err
		}

		card := CardFromName(cards, h[idx])
		game, err := g.play(g.currentPlayer, card)
		if err != nil {
			return g, out, err
		}

		game, err = game.removeFromHand(game.currentPlayer, idx)
		if err != nil {
			return g, out, err
		}

		return game, out, nil
	case "draw":
		g, err := g.drawCard(g.currentPlayer)
		if err != nil {
			return g, out, err
		}
		sayOut("Drew 1 card")
		return g, out, nil
	case "end":
		g, err := g.endTurn()
		if err != nil {
			return g, out, err
		}
		return g, out, nil
	default:
		return g, out, errors.New("Invalid")
	}
}

func GetCardData(data []byte) []Cdata {
	//var rawCards map[string]json.RawMessage
	var rawCards []json.RawMessage
	err := json.Unmarshal(data, &rawCards)
	if err != nil {
		// return err
	}

	cards := make([]Cdata, 30)
	for i, v := range rawCards {
		var c Cdata
		err := json.Unmarshal(v, &c)
		if err != nil {
			// return err
		}

		cards[i] = c
	}
	return cards
}
