package game

import (
	"errors"
	"fmt"
	"slices"
	"math/rand"
)

const (
	CardPerDmg = 2 
	MaxHp = 8
	DisappearRecoil = 2
	AllyRecoil = 1
	PyrusBalioDmg = 1
	EnhanciusBuff = 2
	DragoniusDmg = 3
	AngeliDustioHeal = 2
	VitaliusBuff = 2
	MeteorusDmg = 1
	DracusPyrioDmg = 7
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
	manaMax := s.manaMax + s.numOfPerms("Aquarius")	

	if mana < manaMax {
		p.manaCap++
	}

	mana += p.moreMana
	p.moreMana = 0

	s.Mana = mana

	field := &s.field[s.currentPlayer]
	for i := 0; i < len(*field); i++ {
		(*field)[i].protected = false
	}
	
	if !s.testing {
		s, _ = s.drawCard(p.id)
	}
	numLibrarius := s.numOfPerms("Librarius")
	for i := 0; i < numLibrarius; i++ {
		s, _ = s.drawCard(p.id)
	}

	return s
}

func CardFromName(cards []Cdata, n CardName) Playable {
	c := cards[int(n)]

	if c.Type == "perm" {
		return Perm{
			name: c.Name,
			cost: c.Hp,
			cname: c.CName,
		}
	}

	if c.Type == "instant" {
		return Instant{
			name: c.Name,
			cost: c.Hp, // should be c.Cost 
			cname: c.CName,
		}
	}

	return Card{
		name: c.Name,
		hp:   c.Hp,
		atk0: c.Atk0,
		atk1: c.Atk1,
		cname: c.CName,
	}
}

func (s State) playPerm(player playerID, p Perm) (State, error) {
	s, pTarget, err := s.addPerm(player, p)
	if err != nil {
		return s, err
	}

	switch p.name {
	case "Mortius", "Enhancius", "Vitalius", "Bubublius", "Armorius": // Attachments
		s = s.awaitPerm(pTarget)
	case "Dragonius":
		c := Card{	
			name: "Dragonius",
			hp: 3,
			atk0: Attack{
				Name: "Dragon Breath",
				Dmg: 3,
			},
		}
		s.dragons[player][s.lenPermsOf(player) - 1] = c 
	case "Aquarius", "Conjorius", "Librarius", "Meteorus":
	default:
		return s, errors.New("unexpected perm")
	}
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

	if s.testing == false {
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
	spellsThatAwait := []string{"Pyrus Balio", "Protectio", "Cancelio", "Angeli Dustio", "Dracus Pyrio", "Retrievio", "Extractio"}

	if spell.name == "Dracus Pyrio" &&
	   len(s.players[s.currentPlayer].hand) == 0 {
		return s, GameErr{"Can't play Dracus Pyrio with an empty hand"}
	}

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
	s.output.Println("Hello")
	player.hand = append(player.hand, d[len(d)-1])
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
	for k, v := range s.permanents {
		v.activated = false
		s.permanents[k] = v
	}
	return s.startTurn(), nil
}

func (s State) setMana(n int) State {
	s.Mana = n
	return s
}

func (s State) activatePerm(pt PermTarget) (State, error) {
	p, ok := s.permanents[pt]
	if !ok {
		return s, errors.New("couldnt find perm")
	}

	if p.activated {
		return s, errors.New("Perm alr activated")
	}
	p.activated = true
	s.permanents[pt] = p
	
	if p.name == "Meteorus" {
		t, err := s.randomTarget()
		if err != nil {
			return s, err
		}

		return s.DoDmg(int(t.pID), t.id, MeteorusDmg), nil 
	}
	return s, errors.New("not a valid perm name")
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
		d := PermTarget{pID: atkr.pID, id: atkr.id}
		dragon, ok := s.permanents[d]
		if !ok {
			return s, errors.New("couldnt find dragon")
		}
		if dragon.activated {
			return s, errors.New("attack with dragon only once per turn")
		}
		dragon.activated = true
		s.permanents[d] = dragon
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
	if dmg > 0 {
		if c.protected {
			dmg = 0
		} else {
			if c.resistance {
				// constants
				dmg--
			}
			if c.attached == "Armorius" {
				dmg--
			}
		}
	}

	return g.doRawDmg(c, dmg)
}

func (g State) doRawDmg(c *Card, dmg int) State {
	maxH := MaxHp
	if c.attached == "Vitalius" {
		maxH += VitaliusBuff	
	}

	newHp := c.hp - dmg
	if newHp < 0 {
		newHp = 0
	} else if newHp > maxH {
		newHp = maxH 
	}
	
	if c.hp > 0 && newHp == 0 {
		for i, _ := range g.players {
			g.players[i].moreMana += g.numOfPermsOf(playerID(i), "Conjorius")	
		}
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
		p := s.permanents[PermTarget{t.pID, t.id}]
		if p.name == "Dragonius" {
			return &s.dragons[t.pID][t.id], nil
		}
		return nil, errors.New("perm is not a valid target")
	}
		
	if t.area != Wizard {
		return nil, errors.New("Unexpected area in target")
	}
	return &s.field[t.pID][t.id], nil
}

func (s State) spellTarget(defr target) (State, error) {
	spell, ok := castSpellOnTarget[s.awaiting.spellName]
	if !ok {
		return s, errors.New("Invalid spell name")
	}

	return spell(s, defr)
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
			return s.removePerm(PermTarget{
					pID: defr.pID, id: defr.id})
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

func (s State) awaitPerm(pt PermTarget) State {
	p, ok := s.permanents[pt]
	if !ok {
		panic("COuldnt find perm to await")
	}
	s.awaiting = Await{
		isTrue: true,
		spell: true,
		spellName: p.getName(), 
		perm: pt,
	}
	return s
}

func (s State) cancelAwait() State {
	s.awaiting = Await{}
	return s
}

func (s State) randomTarget() (target, error) {
	targets := []target{}	
	for p := range s.players {
		for i := range MaxFieldLen {
			t := target{pID: playerID(p), id: i}  
			c, err := s.cardFromTarget(t)
			if err != nil {
				continue
			}

			if c.hp != 0 {
				targets = append(targets, t)
			}
		}	
	}
	if len(targets) == 0 {
		return target{}, errors.New("No targets to damage")
	}
	return targets[rand.Intn(len(targets))], nil 
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
