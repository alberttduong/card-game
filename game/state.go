package game

import (
	"errors"
	"fmt"
	"slices"
)

const (
	CardPerDmg = 2 
	MaxHp = 8
	DisappearRecoil = 2
	AllyRecoil = 1
	PyrusBalioDmg = 1
	EnhanciusBuff = 2
	DragoniusDmg = 3
)

func test() {
	fmt.Println("temp")
}

// make useful
func (s State) SetDeck(p int, d ...CardName) State {
	s.players[p].deck = []CardName{Librarian}
	return s
}

func (s State) startTurn() State {
	p := &s.players[s.currentPlayer]

	mana := p.manaCap
	if mana < s.manaMax {
		p.manaCap++
	}

	if p.moreMana {
		mana++
		p.moreMana = false
	}

	s.Mana = mana

	field := &s.field[s.currentPlayer]
	for i := 0; i < len(*field); i++ {
		(*field)[i].protected = false
	}

	return s
}

func CardFromName(cards []Cdata, n CardName) Playable {
	c := cards[int(n)]

	if c.Type == "perm" {
		return Perm{
			name: c.Name,
			cost: c.Hp,
		}
	}

	if c.Type == "instant" {
		return Instant{
			name: c.Name,
			cost: c.Hp, // should be c.Cost 
		}
	}

	return Card{
		name: c.Name,
		hp:   c.Hp,
		atk0: c.Atk0,
		atk1: c.Atk1,
	}
}

func (s State) playPerm(player playerID, p Perm) (State, error) {
	switch p.name {
	case "Mortius", "Enhancius":
		s = s.awaitSpell(p)
	case "Dragonius":
		p.card = Card{	
			name: "Dragonius",
			hp: 3,
			atk0: Attack{
				Name: "Dragon Breath",
				Dmg: 3,
			},
		}
	case "Aquarius":
	default:
		return s, errors.New("unexpected perm")
	}
	s.perms[player] = append(s.perms[player], p) 
	return s, nil
}

func (s State) play(p playerID, c Playable) (State, error) {
	cardType := c.getType()

	if cardType == Wizard {
		if len(s.field[p]) == MaxFieldLen {
			return s, errors.New("Field is at max capacity")
		}
		s.field[p] = append(s.field[p], c.(Card))
		return s, nil
	}

	if s.useMana == true {
		cost := c.getCost()
		if s.players[p].discountSpell {
			cost--
			s.players[p].discountSpell = false
		}
		if s.Mana < cost {
			return s, errors.New("Not enough mana")
		}
		s.Mana -= cost 
	}

	switch cardType {
	case Permanent:
		return s.playPerm(p, c.(Perm))
	case InstantSpell:
		return s.playInstant(c.(Instant))
	}
	return s, nil
}

func (s State) playInstant(spell Instant) (State, error) {
	spellsThatAwait := []string{"Pyrus Balio", "Protectio", "Cancelio"}   

	if slices.Contains(spellsThatAwait, spell.name) {
		return s.awaitSpell(spell), nil
	}
	switch spell.name {
	case "Dralio":
		s, _ = s.drawCard(s.currentPlayer)
		s, err := s.drawCard(s.currentPlayer)
		return s, err	
	}

	return s, errors.New("Invalid instant")
}

func (s State) drawCard(p playerID) (State, error) {
	player := &s.players[p]
	d := player.deck
	if len(d) == 0 {
		return s, errors.New("Deck empty")
	}

	s.players[p].hand = append(s.players[p].hand, d[len(d)-1])
	d = d[:len(d)]
	return s, nil
}

func (s State) removeFromHand(p playerID, idx int) (State, error) {
	hand := s.players[p].hand
	if idx >= len(hand) || idx < 0 {
		return s, errors.New("Index out of bounds")
	}
	hand[idx] = hand[len(hand)-1]
	hand = hand[:len(hand)-1]
	s.players[p].hand = hand
	return s, nil
}

func (s State) addToDiscard(p playerID, c CardName) State {
	s.discard[p] = append(s.discard[p], c)
	return s
}

type target struct {
	pID    playerID
	area cardType
	id     int
	atkNum int
}

func (s State) endTurn() (State, error) {
	s.currentPlayer = playerID(int(s.currentPlayer+1) % s.numPlayers)
	for _, permField := range s.perms {
		for i, _ := range permField {
			permField[i].activated = false	
		}
	}
	return s.startTurn(), nil
}

func (s State) setMana(n int) State {
	s.Mana = n
	return s
}

func (s State) attack(atkr, defr target) (State, error) {
	atkrCard, err := s.cardFromTarget(atkr)
	if err != nil { 
		return s, err
	}

	defrCard, err := s.cardFromTarget(defr)
	if err != nil { 
		return s, err
	}

	atk, err := atkrCard.atk(atkr.atkNum)
	if errors.Is(err, DragoniusAtkErr{}) {
		dragon, _ := s.permFromTarget(atkr)
		if dragon.activated == true {
			return s, errors.New("attack with dragon only once per turn")
		}
		dragon.activated = true
		return s.setAwait(atkr), nil 
	}
	if err != nil {
		return s, err
	}


	sideEffect, ok := atkSideEffects[atk.Name]	
	if ok {
		s = sideEffect(s, atkr, defr)
	}

	dmg := s.baseDamage(atkr, atk)

	if atk.Name == "bypass" {
	 	s = s.doRawDmg(defrCard, dmg)
		return s, nil
	}

	defrWasAlive := defrCard.hp > 0

	s = s.DoDmgToCard(defrCard, dmg)

	if defrWasAlive && defrCard.hp == 0 {
		if atk.Name == "frenzy" {
			s = s.setAwait(atkr)
		}
		if defrCard.attached == "Mortius" {
			s = s.DoDmgToCard(atkrCard, 2)
		}
	}

	return s, nil
}

func (s State) allAllies(f func(*Card) bool, t target) bool {
	for i := 0; i < MaxFieldLen - 1; i++ {
		fId := (t.id + 1 + i) % MaxFieldLen 
		card, err := s.cardFromTarget(target{pID: t.pID, id: fId})
		if err != nil {
			continue
		}
		if f(card) == false {
			return false
		}
	}
	return true
}

func (s State) applyToAllies(f func(*Card), t target) {
	for i := 0; i < MaxFieldLen - 1; i++ {
		fId := (t.id + 1 + i) % MaxFieldLen
		card, err := s.cardFromTarget(target{pID: t.pID, id: fId})
		if err != nil {
			continue
		}	
		f(card)
	}
}

func (g State) DoDmg(p, id, dmg int) State {
	return g.DoDmgToCard(&g.field[p][id], dmg)
}

func (g State) DoDmgToCard(c *Card, dmg int) State {
	if c.protected {
		dmg = 0
	}

	if dmg > 0 && c.resistance == true {
		dmg--
	}

	return g.doRawDmg(c, dmg)
}

func (g State) doRawDmg(c *Card, dmg int) State {
	newHp := c.hp - dmg
	if newHp < 0 {
		newHp = 0
	} else if newHp > MaxHp {
		newHp = MaxHp 
	}
	c.hp = newHp
	return g
}

func (s State) expectedTargetType() cardType {
	atkr := s.awaiting.atkr
	atkrCard := &s.field[atkr.pID][atkr.id]

	name := atkrCard.atk0.Name
	if atkr.atkNum == 1 {
		name = atkrCard.atk1.Name
	}

	targetsPerm := []string{
		"removePerm",
	}

	if slices.Contains(targetsPerm, name) {
		return Permanent
	}

	return Wizard 
}
	
func (s State) cardFromTarget(t target) (*Card, error) {
	if err := s.checkTarget(t); err != nil {
		return nil, err
	}
	if t.area == Permanent {
		p := &s.perms[t.pID][t.id]
		if p.name == "Dragonius" {
			return &p.card, nil
		}
		return nil, errors.New("perm is not a valid target")
	}
		
	if t.area != Wizard {
		return nil, errors.New("Unexpected area in target")
	}
	return &s.field[t.pID][t.id], nil
}

func (s State) permFromTarget(t target) (*Perm, error) {
	if err := s.checkTarget(t); err != nil {
		return nil, err
	}
	if t.area != Permanent {
		return nil, errors.New("Target not a valid perm")
	}
	return &s.perms[t.pID][t.id], nil
}

func (s State) spellTarget(defr target) (State, error) {
	switch s.awaiting.spellName {
	case "Pyrus Balio":
		return s.DoDmg(int(defr.pID), defr.id, PyrusBalioDmg), nil
	case "Protectio":
		c, err := s.cardFromTarget(defr)	
		if err != nil {
			return s, err
		}
		c.protected = true 
		return s, nil 
	// Attach
	case "Mortius", "Enhancius":
		c, err := s.cardFromTarget(defr)	
		if err != nil {
			return s, err
		}
		c.attached = s.awaiting.spellName 
		return s, nil
	case "Cancelio":
		p, err := s.permFromTarget(defr)
		if err != nil {
			return s, err
		}
		*p = Perm{} 
		return s, nil
	}
	return s, errors.New("awaiting invalid spell name")
}

func (s State) target(defr target) (State, error) {
	if !s.awaiting.isTrue {
		return s, errors.New("Not awaiting target")
	}

	if s.awaiting.spell {
		return s.spellTarget(defr)
	}

	targetType := s.expectedTargetType()  
	if defr.area != targetType {
		return s, errors.New("Unexpected target")
	}
	if err := s.checkTarget(defr); err != nil {
		return s, err
	}


	atkr := s.awaiting.atkr
	atkrCard, err := s.cardFromTarget(atkr) 
	if err != nil {
		return s, err
	}

	defrCard, wrongTargetError := s.cardFromTarget(defr) 
	atk, err := atkrCard.atk(atkr.atkNum)

	if errors.Is(err, DragoniusAtkErr{}) {
		if wrongTargetError != nil {
			return s, errors.New("invalid target")
		}
		s = s.cancelAwait()
		return s.DoDmgToCard(defrCard, DragoniusDmg), nil
	}
	if err != nil {
		return s, err
	}

	if targetType == Permanent {
		switch atk.Name {
		case "removePerm":
			s.perms[defr.pID][defr.id] = Perm{} 
		}
		return s, errors.New("Unexpected attack")
	}


	switch atk.Name {
	case "revive":
		atkrCard.hp = 0
		defrCard.hp = MaxHp 
	case "attackTwice":
		s, err := s.attack(atkr, defr) 
		s = s.cancelAwait()
		return s, err
	case "frenzy":
		defrWasAlive := defrCard.hp > 0

		s, err := s.attack(atkr, defr)	
		if err != nil {
			return s, nil
		}

		if defrWasAlive && defrCard.hp == 0 {
			return s, nil
		}	
	}

	s = s.cancelAwait()
	return s, nil
}

func (s State) setAwait(a target) State {
	s.awaiting = Await{
		isTrue: true,
		atkr:   a,
	}
	return s
}

func (s State) awaitSpell(spell Playable) State {
	s.awaiting = Await{
		isTrue: true,
		spell: true,
		spellName: spell.getName(), 
	}
	return s
}

func (s State) cancelAwait() State {
	s.awaiting = Await{}
	return s
}

func (c Card) atk(n int) (Attack, error) {
	if c.name == "Dragonius" {
		return Attack{}, DragoniusAtkErr{} 
	}
	
	if n != 0 && n != 1 {
		panic("invalid atk number")
	}
	if n == 1 {
		return c.atk1, nil
	}
	return c.atk0, nil
}
