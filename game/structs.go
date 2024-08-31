package game

import (
	"errors"
	"fmt"
)

type CardName int

// Order matters
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
	Protectio
	PyrusBalio
	Mortius
	Enhancius
	Dragonius
	Cancelio
	Conjorius
	AngeliDustio
	Vitalio
	Dralio
	Librarius
	Aquarius
	Bubublius
	Meteorus
	Armorius
	DracusPyrio
	Retrievio
	Extractio

	MaxFieldLen = 3
	MaxPermLen  = 7
	MaxPlayers  = 5
)

type Attack struct {
	Name string `json:"name"`
	Dmg  int    `json:"dmg"`
}

type Cdata struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Hp   int    `json:"hp"`
	Atk0 Attack `json:"atk1"`
	Atk1 Attack `json:"atk2"`
}

type cardType int

type Playable interface {
	getType() cardType
	getCost() int
	getName() string
}

const (
	Wizard cardType = iota
	Permanent
	InstantSpell
)

// Should be named Wizard
type Card struct {
	name string
	hp   int
	atk0 Attack
	atk1 Attack

	protected  bool
	resistance bool
	attached   string
}

func (c Card) getType() cardType {
	return Wizard 
}

func (c Card) getCost() int {
	return 0 
}

func (c Card) getName() string {
	return c.name 
}

type Perm struct {
	name string
	cost int
	card Card
	activated bool
}

func (p Perm) getType() cardType {
	return Permanent 
}

func (p Perm) getCost() int {
	return p.cost 
}

func (p Perm) getName() string {
	return p.name 
}

type Instant struct {
	name string
	cost int
}

func (i Instant) getType() cardType {
	return InstantSpell 
}

func (i Instant) getCost() int {
	return i.cost 
}

func (i Instant) getName() string {
	return i.name 
}

type playerID int

type Player struct {
	id      playerID
	deck    []CardName
	hand    []CardName
	manaCap int

	magicianHealth int
	moreMana       bool
	discountSpell  bool
}

func InitPlayer(p playerID) Player {
	return Player{
		id:      p,
		deck:    []CardName{},
		hand:    make([]CardName, 0, 7),
		manaCap: 1,
	}
}

func (p Player) String() string {
	return fmt.Sprintf("Player %d\n\tDeck: %v\n\tHand: %v",
		p.id, p.deck, p.hand)
}

type Await struct {
	isTrue bool
	atkr   target
	spell bool
	spellName string 
}

type State struct {
	numPlayers    int
	players       [MaxPlayers]Player
	currentPlayer playerID
	discard       [MaxPlayers][]CardName
	field         [MaxPlayers][]Card
	perms         [MaxPlayers][]Perm
	
	Mana          int
	manaMax       int
	useMana bool

	awaiting Await
}

func initArea[T Playable](pType T, numPlayers int) [MaxPlayers][]T {
	var maxLen int
	switch pType.getType() {
	case Wizard:
		maxLen = MaxFieldLen
	case Permanent:
		maxLen = MaxPermLen
	}

	area := [MaxPlayers][]T{}
	for i := 0; i < numPlayers; i++ {
		area[i] = make([]T, 0, maxLen)
	}
	return area
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
		numPlayers:    players,
		currentPlayer: 0,
		discard:       [MaxPlayers][]CardName{},
		field:         initArea(Card{}, players),
		perms:   initArea(Perm{}, players),
		manaMax: 6,
		useMana: false,
	}
	return s.startTurn(), nil
}

func InitStateUsingMana(players int) (State, error) {
	s, err := InitState(players)
	if err != nil {
		return State{}, err
	}
	s.useMana = true
	return s, nil
}

func (s State) String() string {
	return fmt.Sprintf(
		"Await: %v\n"+
			"Current Player: %d\n"+
			"mana: %d\n"+
			"%v\n%v\n"+
			"Fields: %v\nDiscard: %v",
		s.awaiting,
		s.currentPlayer,
		s.Mana,
		s.players[0],
		s.players[1],
		s.field,
		s.discard,
	)
}
