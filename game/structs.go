package game 

import (
"fmt"
	"errors"
)

type CardName int

//Order matters
//go:generate stringer -type CardName
const (
	Librarian CardName = iota
	Magician
	Shieldmancer
	MindMage
	Angel
	Pyromancer
	Bloodeater
	Conjurer
	Mortician
	
	MaxFieldLen = 3
	MaxPlayers = 5
)

type Attack struct {
	Name string `json:"name"`
	Dmg int `json:"dmg"`
}

type Cdata struct {
	Name string `json:"name"`
	Hp int `json:"hp"`
	Atk1 Attack `json:"atk1"`	
	Atk2 Attack `json:"atk2"`	
}

type Card struct {
	name string
	hp     int
	atk1 Attack
	atk2 Attack
}

func IsValidCardId(n int, cards []Cdata) bool {
	return n > 0 && n < len(cards)
}

func CardFromName(cards []Cdata, n CardName) Card {
	c := cards[int(n)]
	return Card {
		name: c.Name,
		hp: c.Hp,
		atk1: c.Atk1,
		atk2: c.Atk2,
	}
}
		
type playerID int

type Player struct {
	id playerID
	deck []CardName
	hand []CardName
	manaCap int
}

func InitPlayer(p playerID) Player {
	return Player{
		id:   p,
		deck: []CardName{},
		hand: make([]CardName, 0, 7),
		manaCap: 1,
	}
}

func (p Player) String() string {
	return fmt.Sprintf("Player %d\n\tDeck: %v\n\tHand: %v",
		p.id, p.deck, p.hand)
}

type State struct {
	numPlayers    int
players       [MaxPlayers]Player
	currentPlayer playerID
	discard       [MaxPlayers][]CardName
	field         [MaxPlayers][]Card
	Mana int
	manaMax int
}

func InitState(players int) (State, error) {
	if players < 2 || players > MaxPlayers {
		return State{}, errors.New("Invalid number of players")
	}

	s := State{
		players: [MaxPlayers]Player{
			InitPlayer(0),
			InitPlayer(1),
		},
		numPlayers: players,
		currentPlayer: 0,
		discard:       [MaxPlayers][]CardName{},
		field: [MaxPlayers][]Card{
			make([]Card, 0, MaxFieldLen),
			make([]Card, 0, MaxFieldLen),
			make([]Card, 0, MaxFieldLen),
			make([]Card, 0, MaxFieldLen),
		},
		manaMax: 6,
	}
	return s.startTurn(), nil
}

func (s State) String() string {
	return fmt.Sprintf(
		"Current Player: %d\n" +
		"mana: %d\n" +
		"%v\n%v\n" + 
		"Fields: %v\nDiscard: %v",
		s.currentPlayer,
		s.Mana,
		s.players[0],
		s.players[1],
		s.field,
		s.discard,
	)
}

func (s State) Title() string {
	return fmt.Sprintf("%d", s.currentPlayer)
}
