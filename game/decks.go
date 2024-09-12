package game

import (
	"bytes"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

const (
	MaxDeckLength = 20
	MaxWizards    = 3
	MaxCopies     = 4
)

type DeckError struct{ msg string }

func (d DeckError) Error() string {
	return d.msg
}

var (
	DeckLineErr           = DeckError{"Must be 2 numbers on each line"}
	DeckFormatErr         = DeckError{"Format is wrong"}
	MaxCardCopiesErr      = DeckError{fmt.Sprintf("Over max copies per card, %d", MaxCopies)}
	MaxDeckErr            = DeckError{fmt.Sprintf("Over deck limit, %d", MaxDeckLength)}
	InvalidCardErr        = DeckError{"Invalid card number"}
	LimitOneEachWizardErr = DeckError{"Limit one of each wizards"}
)

func ParseDeck(data []byte) (map[int]int, error) {
	d := bytes.NewBuffer(data).String()

	deck := make(map[int]int)

	lines := strings.Split(d, "\n")
	// dont know why we have to subtract 2 but we do
	for i := 0; i < len(lines) - 1; i++ {
		nums := strings.Split(lines[i], " ")
		if len(nums) != 2 {
			return deck, DeckFormatErr 
		}

		amount, err := strconv.Atoi(nums[0])
		if err != nil {
			return deck, DeckFormatErr 
		}

		id, err := strconv.Atoi(nums[1])
		if err != nil {
			return deck, DeckFormatErr 
		}
		deck[id] = amount
	}
	return deck, nil
}

func isWizard(cards []Cdata, id int) bool {
	return cards[id - 1].Type == "wizard"
}

type DeckEntry struct {
	ID, Amount int
}

func SortEntries(list []DeckEntry) []DeckEntry {
	list = slices.DeleteFunc(list, func(a DeckEntry) bool {
		return a.Amount < 1 || a.ID < 1
	})
	slices.SortStableFunc(list, func(a, b DeckEntry) int {
		return a.ID - b.ID
	})
	return list
}

func SortedDeckList(cards []Cdata, m map[int]int) (list []DeckEntry) {
	for k, v := range m {
		list = append(list, DeckEntry{k, v})
	}
	SortEntries(list)

	return
}

func ValidateDeck(cards []Cdata, deck map[int]int) error {
	total := 0
	totalWizards := 0
	for id, amount := range deck {
		if id < 0 || id >= len(cards) {
			return InvalidCardErr
		}

		if amount > 4 {
			return MaxCardCopiesErr
		}

		total += amount
		if total > MaxDeckLength {
			return MaxDeckErr
		}

		if isWizard(cards, id) {
			if amount != 1 {
				return LimitOneEachWizardErr
			}
			totalWizards++
		}
	}
	if totalWizards != MaxWizards {
		return DeckError{fmt.Sprintf("Expected 3 wizards got %d", totalWizards)} 
	}
	return nil
}

func EntriesToBytes(entries []DeckEntry) []byte {
	s := ""
	for _, e := range entries {
		s += fmt.Sprintf("%d %d\n", e.Amount, e.ID)	
	}
	return []byte(s)
}

func SearchSortedEntries(entries []DeckEntry, id int) (int, bool) {
	return slices.BinarySearchFunc(entries, DeckEntry{ID: id}, func (a, b DeckEntry) int {
		return a.ID - b.ID
	})
}

func (s State) SetDeckFromMap(p int, cards []Cdata, d map[int]int) (State, error) {
	s.Players[p].deck = []CardName{}
	if err := ValidateDeck(cards, d); err != nil {
		return s, err
	}

	for id, amount := range d {
		s.Players[p].deck = slices.Concat(
			s.Players[p].deck,
			slices.Repeat([]CardName{CardName(id)}, amount))
	}

	return s, nil
}
