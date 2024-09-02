package game

import (
	"errors"
	"fmt"
	"log"
	"bytes"
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
}

type Cdata struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Hp   int    `json:"hp"`
	Atk0 Attack `json:"atk1"`
	Atk1 Attack `json:"atk2"`
	CName CardName
}

type cardType int

type Playable interface {
	getType() cardType
	getCost() int
	getName() string
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
	name string
	cname CardName
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

func (c Card) getCardName() CardName {
	return c.cname 
}

func (c Card) getCost() int {
	return 0 
}

func (c Card) getName() string {
	return c.name 
}

type Perm struct {
	name string
	cname CardName
	cost int
	card Card
	activated bool
	attachedTo target
}

func (p Perm) String() string {
	return fmt.Sprintf("%v\nActivated: %v", p.name, p.activated)
}

func (p Perm) getType() cardType {
	return Permanent 
}

func (c Perm) getCardName() CardName {
	return c.cname 
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
	cname CardName
}

func (i Instant) getType() cardType {
	return InstantSpell 
}

func (i Instant) getCardName() CardName {
	return i.cname
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
	moreMana       int 
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
	perm PermTarget 
}

type State struct {
	numPlayers    int
	players       [MaxPlayers]Player
	currentPlayer playerID
	discard       [MaxPlayers][]CardName
	field         [MaxPlayers][]Card
	permanents    map[PermTarget]Perm 
	dragons       [MaxPlayers][]Card
	
	Mana          int
	manaMax       int
	useMana bool 
	testing bool

	awaiting Await
	logs     *bytes.Buffer 
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
		players: [MaxPlayers]Player{
			InitPlayer(0),
			InitPlayer(1),
		},
		numPlayers:    players,
		currentPlayer: 0,
		discard:       [MaxPlayers][]CardName{},
		field:         initArea(Card{}, players),
		dragons:       initDragons(players), 
		permanents: make(map[PermTarget]Perm),
		manaMax: 6,
		logs: &buf,
	}
	s.output = log.New(&buf, "Game: ", log.Lmsgprefix)

	return s.startTurn(), nil
}


func NewTestGame(players int) (State, error) {
	s, err := NewGame(players)
	if err != nil {
		return State{}, err
	}
	s.testing = true
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

func lenPerm(p [MaxPermLen]Perm) (n int) {
	empty := Perm{}
	for _, perm := range p {
		if perm != empty {
			n++
		}
	}
	return
}

func (s State) numOfPerms(name string) (n int) {
	for _, p := range s.permanents {
		if p.name == name {
			n++
		}
	}
	return
}

func (s State) lenPermsOf(p playerID) (n int) {
	for i, _ := range s.permanents {
		if i.pID == p {
			n++
		}
	}
	return
}

func (s State) numOfPermsOf(p playerID, name string) (n int) {
	for i := 0; i < MaxPermLen; i++ { 
		p, ok := s.permanents[PermTarget{pID: p, id: i}]
		if ok && p.name == name {
			n++
		}
	}
	return
}

func (s State) numOfMyPerms(name string) (n int) {
	return s.numOfPermsOf(s.currentPlayer, name)
}

var EmptyPerm = Perm{}

type PermTarget struct {
	pID playerID 
	id int
}
func (s State) addPerm(id playerID, p Perm) (State, PermTarget, error) {
	for i := 0; i < MaxPermLen; i++ { 
		t := PermTarget{pID: id, id: i}	
		_, ok := s.permanents[t] 
		if !ok {
			s.permanents[t] = p	
			return s, t, nil
		}
	}
	return s, PermTarget{}, errors.New("Max number of perms reached")
}

func (s State) removePerm(pt PermTarget) (State, error) {
	p, ok := s.permanents[pt]
	if !ok {
		return s, errors.New("Couldnt find perm to remove")
	}

	delete(s.permanents, pt) 

	c, err := s.cardFromTarget(p.attachedTo)
	if err != nil {
		return s, nil
	}

	c.attached = ""
	if p.name == "Vitalius" {
		s = s.doRawDmg(c, 2)
	}

	return s, nil
}
