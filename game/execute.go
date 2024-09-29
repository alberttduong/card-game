package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

func isValidCardNumber(n int) error {
	if n < 1 || n > 30 {
		return errors.New("Invalid card number")
	}
	return nil
}

func convertArgs(n int, args ...string) ([]int, error) {
	lenArgs := len(args)
	if lenArgs < n {
		return []int{},
			InputErr{fmt.Sprintf("Expected %d args, got %d",
				n, lenArgs)}
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

func (g State) Execute(cards []Cdata, args ...string) (State, error) {
	switch args[0] {
	case "target":
		nums, err := convertArgs(2, args[1:]...)
		if err != nil {
			return g, err
		}

		g, err = g.target(target{pID: playerID(nums[0]), id: nums[1]})
		if err != nil {
			return g, err
		}

		return g, nil
	case "targetperm":
		nums, err := convertArgs(2, args[1:]...)
		if err != nil {
			return g, err
		}

		g, err = g.target(target{area: Permanent, 
					pID: playerID(nums[0]), id: nums[1]})
		if err != nil {
			return g, err
		}

		return g, nil
	case "create":
		nums, err := convertArgs(1, args[1:]...)
		if err != nil {
			return g, err
		}

		if err := isValidCardNumber(nums[0]); err != nil {
			return g, err
		}

		game, err := g.play(g.CurrentPlayer, CardFromName(cards, CardName(nums[0])))
		if err != nil {
			return g, err
		}

		return game, nil
	case "setmana":
		nums, err := convertArgs(1, args[1:]...)
		if err != nil {
			return g, err
		}

		return g.setMana(nums[0]), nil
	case "attack":
		nums, err := convertArgs(5, args[1:]...)
		if err != nil {
			return g, err
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
			return g, err
		}

		return newG, nil
	case "atk":
		nums, err := convertArgs(7, args[1:]...)
		if err != nil {
			return g, err
		}

		newG, err := g.attack(
			target{
				area:   cardType(nums[0]),
				pID:    playerID(nums[1]),
				id:     nums[2],
				atkNum: nums[3],
			},
			target{
				area:   cardType(nums[4]),
				pID: playerID(nums[5]),
				id:  nums[6],
			})

		if err != nil {
			return g, err
		}

		return newG, nil
	case "play":
		nums, err := convertArgs(1, args[1:]...)
		if err != nil {
			return g, err
		}

		idx := nums[0]
		h := g.Players[g.CurrentPlayer].Hand
		if idx < 0 || idx > len(h)-1 {
			return g, err
		}

		card := CardFromName(cards, h[idx])
		game, err := g.play(g.CurrentPlayer, card)
		if err != nil {
			return g, err
		}

		game, err = game.removeFromHand(game.CurrentPlayer, idx)
		if err != nil {
			return g, err
		}

		return game, nil
	case "draw":
		g, err := g.drawCard(g.CurrentPlayer)
		if err != nil {
			return g, err
		}
		return g, nil
	case "showdeck":
		return g.showDeck(g.CurrentPlayer), nil
	case "activate":
		//todo
		nums, err := convertArgs(2, args[1:]...)
		if err != nil {
			return g, err
		}

		g, err = g.activatePerm(PermTarget{playerID(nums[0]), nums[1]})
		return g, err
	case "end":
		g, err := g.endTurn()
		if err != nil {
			return g, err
		}
		return g, nil
	default:
		return g, errors.New("Invalid Command")
	}
}

func GetCardData(data []byte) []Cdata {
	var rawCards []json.RawMessage
	err := json.Unmarshal(data, &rawCards)
	if err != nil {
		panic("Couldn't parse card data")
	}

	cards := make([]Cdata, 30)
	for i, v := range rawCards {
		var c Cdata
		err := json.Unmarshal(v, &c)
		if err != nil {
			panic("Couldn't parse card data")
		}

		c.CName = CardName(i + 1)
		cards[i] = c
	}

	return cards
}
