package game 

import (
	"fmt"
	"errors"
)

type Cdata struct {
	Hp int `json:"hp"`
	Atk int `json:"atk"`
}


type cardName string 
const (
	fireball cardName = "fireball"
	wizard = "wizard"
	goblin = "goblin"
	poison = "poison"

	MaxFieldLen = 3
	MaxPlayers = 5
)

func CardFromName(d map[string]Cdata, n cardName) Card {
	return Card {
		hp: 1,
		atk: 2,
	}
}
		
type Card struct {
	player playerID
	hp     int
	atk    int
}

type playerID int

type Player struct {
	id playerID
	deck []cardName
	hand []cardName
}

func InitPlayer(p playerID) Player {
	return Player{
		id:   p,
		deck: []cardName{wizard, wizard},
		hand: make([]cardName, 0, 7),
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
	discard       [MaxPlayers][]cardName
	field         [MaxPlayers][]Card
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
		discard:       [MaxPlayers][]cardName{},
		field: [MaxPlayers][]Card{
			make([]Card, 0, MaxFieldLen),
			make([]Card, 0, MaxFieldLen),
			make([]Card, 0, MaxFieldLen),
			make([]Card, 0, MaxFieldLen),
		},
	}
	return s, nil
}

func (s State) String() string {
	return fmt.Sprintf(
		"Current Player: %d\n" +
		"%v\n%v\n" + 
		"Fields: %v\nDiscard: %v",
		s.currentPlayer,
		s.players[0],
		s.players[1],
		s.field,
		s.discard,
	)
}

func (s State) Title() string {
	return fmt.Sprintf("%d", s.currentPlayer)
}
