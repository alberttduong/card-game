package game

import (
	"bytes"
	"errors"
	"fmt"
	"log"
)

type CardName int

// Order matters
//
//go:generate stringer -type=CardName
const (
	None CardName = iota
	Librarian
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
	Vitalius
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
	Desc string `json:"desc"`
}

type Cdata struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Desc  string `json:"desc"`
	Hp    int    `json:"hp"`
	Atk0  Attack `json:"atk1"`
	Atk1  Attack `json:"atk2"`
	CName CardName
}

type cardType int

type Playable interface {
	getType() cardType
	getCost() int
	getCardName() CardName
}

const (
	Wizard cardType = iota
	Permanent
	InstantSpell
	Deck
)

// Should be named Wizard
type Card struct {
	HP    int
	CName CardName
	Atk0  Attack
	Atk1  Attack

	//export
	protected  bool
	resistance bool
	attached   CardName
}

func (c Card) getType() cardType {
	return Wizard
}

func (c Card) getCardName() CardName {
	return c.CName
}

func (c Card) getCost() int {
	return 0
}

type Perm struct {
	CName      CardName
	Cost       int
	Activated  bool
	AttachedTo target

	card Card
}

func (p Perm) getType() cardType {
	return Permanent
}

func (c Perm) getCardName() CardName {
	return c.CName
}

func (p Perm) getCost() int {
	return p.Cost
}

type Instant struct {
	CName CardName
	Cost  int
}

func (i Instant) getType() cardType {
	return InstantSpell
}

func (i Instant) getCardName() CardName {
	return i.CName
}

func (i Instant) getCost() int {
	return i.Cost
}

type playerID int

type Player struct {
	ID   playerID
	Hand []CardName
	deck []CardName

	manaCap        int
	magicianHealth int
	moreMana       int
	discountSpell  bool
}

func InitPlayer(p playerID) Player {
	return Player{
		ID:      p,
		deck:    []CardName{},
		Hand:    make([]CardName, 0, 7),
		manaCap: 1,
	}
}

func (p Player) String() string {
	return fmt.Sprintf("Player %d\n\tDeck: %v\n\tHand: %v",
		p.ID, p.deck, p.Hand)
}

type Await struct {
	isTrue    bool
	atkr      target
	spell     bool
	spellName CardName
	perm      PermTarget
}

type State struct {
	NumPlayers    int
	Players       [MaxPlayers]Player
	CurrentPlayer playerID
	Field         [MaxPlayers][]Card
	Permanents    map[PermTarget]Perm
	Dragons       [MaxPlayers][]Card
	Mana          int

	manaMax int
	useMana bool
	testing bool

	awaiting Await
	Logs     *bytes.Buffer
	output   *log.Logger
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

func initDragons(players int) [MaxPlayers][]Card {
	var d [MaxPlayers][]Card
	for i := 0; i < MaxPlayers; i++ {
		d[i] = make([]Card, MaxPermLen)
	}
	return d
}

func NewGame(players int) (State, error) {
	if players < 2 || players > MaxPlayers {
		return State{}, errors.New("Invalid number of players")
	}
	var buf bytes.Buffer

	s := State{
		Players: [MaxPlayers]Player{
			InitPlayer(0),
			InitPlayer(1),
		},
		testing: false,
		NumPlayers:    players,
		CurrentPlayer: 0,
		Field:         initArea(Card{}, players),
		Dragons:       initDragons(players),
		Permanents:    make(map[PermTarget]Perm),
		manaMax:       6,
		Logs:          &buf,
		output:        log.New(&buf, "-", log.Lmsgprefix),
	}

	return s, nil
}

func NewTestGame(players int) (State, error) {
	s, err := NewGame(players)
	if err != nil {
		return State{}, err
	}
	s.testing = true
	return s.startTurn(), nil
}

func (s State) String() string {
	return fmt.Sprintf(
		"Await: %v\n"+
			"Current Player: %d\n"+
			"mana: %d\n"+
			"%v\n%v\n"+
			"Fields: %v\n",
		s.awaiting,
		s.CurrentPlayer,
		s.Mana,
		s.Players[0],
		s.Players[1],
		s.Field,
	)
}

func lenPerm(p [MaxPermLen]Perm) (n int) {
	empty := Perm{}
	for _, perm := range p {
		if perm != empty {
			n++
		}
	}
	return
}

func (s State) numOfPerms(name CardName) (n int) {
	for _, p := range s.Permanents {
		if p.CName == name {
			n++
		}
	}
	return
}

func (s State) lenPermsOf(p playerID) (n int) {
	for i, _ := range s.Permanents {
		if i.pID == p {
			n++
		}
	}
	return
}

func (s State) numOfPermsOf(p playerID, c CardName) (n int) {
	for i := 0; i < MaxPermLen; i++ {
		p, ok := s.Permanents[PermTarget{p, i}]
		if ok && p.CName == c {
			n++
		}
	}
	return
}

func (s State) numOfMyPerms(c CardName) (n int) {
	return s.numOfPermsOf(s.CurrentPlayer, c)
}

var EmptyPerm = Perm{}

type PermTarget struct {
	pID playerID
	id  int
}

func (s State) addPerm(id playerID, p Perm) (State, PermTarget, error) {
	for i := 0; i < MaxPermLen; i++ {
		t := PermTarget{id, i}
		_, ok := s.Permanents[t]
		if !ok {
			s.Permanents[t] = p
			return s, t, nil
		}
	}
	return s, PermTarget{}, errors.New("Max number of perms reached")
}

func (s State) removePerm(pt PermTarget) (State, error) {
	p, ok := s.Permanents[pt]
	if !ok {
		return s, errors.New("Couldnt find perm to remove")
	}

	delete(s.Permanents, pt)

	c, err := s.cardFromTarget(p.AttachedTo)
	if err != nil {
		return s, nil
	}

	c.attached = None
	if p.CName == Vitalius {
		s = s.doRawDmg(c, 2)
	}

	return s, nil
}
