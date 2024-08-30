package game

import (
	"errors"
	"fmt"
)

const (
	CardPerDmg = 2 
	MaxHp = 8
	DisappearRecoil = 2
	AllyRecoil = 1
	PyrusBalioDmg = 1
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
		s.perms[p] = append(s.perms[p], c.(Perm))
	case InstantSpell:
		return s.playInstant(c.(Instant))
	}
	return s, nil
}

func (s State) playInstant(spell Instant) (State, error) {
	switch spell.name {
	case "Pyrus Balio":
		return s.awaitSpell(spell), nil
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
	return s.startTurn(), nil
}

func (s State) setMana(n int) State {
	s.Mana = n
	return s
}

func (s State) attack(atkr target, defr target) (State, error) {
	if err := s.checkTarget(atkr); err != nil {
		return s, err
	}
	if err := s.checkTarget(defr); err != nil {
		return s, err
	}

	atkrCard := s.field[atkr.pID][atkr.id]
	defrCard := &s.field[defr.pID][defr.id]
	atk := atkrCard.atk0
	if atkr.atkNum == 1 {
		atk = atkrCard.atk1
	}

	dmg := atk.Dmg
	switch atk.Name {
	case "draw1":
		s, _ = s.drawCard(s.currentPlayer)
	case "dmgPerCard":
		dmg = len(s.players[atkr.pID].hand) / CardPerDmg 
	case "heal":
		// maxHP
		if defrCard.hp == MaxHp {
			dmg = 0
		}
	case "revive":
		s = s.setAwait(atkr)
	case "disappear":
		s.players[s.currentPlayer].magicianHealth = atkrCard.hp
		h := s.players[s.currentPlayer].hand
		s.players[s.currentPlayer].hand = append(h, Magician)
	case "disappear2":
		newHp := atkrCard.hp - DisappearRecoil 
		s.players[s.currentPlayer].magicianHealth = newHp
		if newHp > 0 {
			h := s.players[s.currentPlayer].hand
			s.players[s.currentPlayer].hand = append(h, Magician)
		}
	case "splash":
		s = s.DoDmg(int(defr.pID), (defr.id+1)%3, dmg)
		s = s.DoDmg(int(defr.pID), (defr.id+2)%3, dmg)
	case "megaSplash":
		for p := 0; p < s.numPlayers; p++ {
			for i := 0; i < 3; i++ {
				if p == int(atkr.pID) {
					continue
				}

				if s.checkTarget(
					target{
						pID: playerID(p),
						id:  i,
					}) != nil {
					continue
				}

				s = s.DoDmg(p, i, 1)
			}
		}
	case "protect":
		protect := func(c *Card) {
			c.protected = true
		}

		s.applyToAllies(protect, atkr)
	case "reduce":
		setResistance := func(c *Card) {
			c.resistance = true
		}
		s.applyToAllies(setResistance, atkr)
	case "moreMana":
		s.players[atkr.pID].moreMana = true
	case "nextDiscount":
		s.players[atkr.pID].discountSpell = true
	case "hurtAllies":
		hurt := func(c *Card) {
			s.DoDmgToCard(c, AllyRecoil)
		}
		s.applyToAllies(hurt, atkr)
	case "double":
		areDead := func(c *Card) bool {
			return c.hp < 1
		}
		if s.allAllies(areDead, atkr) {
			dmg *= 2
		}
	case "bypass":
		s = s.doRawDmg(defrCard, dmg)
		return s, nil
	case "removePerm":
		return s.setAwait(atkr), nil 
	case "attackTwice":
		s = s.setAwait(atkr)
	}

	defrWasAlive := defrCard.hp > 0

	s = s.DoDmg(int(defr.pID), defr.id, dmg)

	if atk.Name == "frenzy" && defrWasAlive && defrCard.hp == 0 {
		s = s.setAwait(atkr)
	}

	return s, nil
}

func (s State) allAllies(f func(*Card) bool, t target) bool {
	for i := 0; i < MaxFieldLen - 1; i++ {
		fId := (t.id + 1 + i) % MaxFieldLen 
		if s.checkTarget(target{pID: t.pID, id: fId}) != nil {
			continue
		}
		if f(&s.field[t.pID][fId]) == false {
			return false
		}
	}
	return true
}

func (s State) applyToAllies(f func(*Card), t target) {
	for i := 0; i < 2; i++ {
		fId := (t.id + 1 + i) % 3
		if s.checkTarget(target{pID: t.pID, id: fId}) != nil {
			continue
		}
		f(&s.field[t.pID][fId])
	}
}

func (g State) DoDmg(p int, id int, dmg int) State {
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

	for _, v := range targetsPerm {
		if name == v {
			return Permanent 
		}
	}

	return Wizard 
}
	

func (s State) spellTarget(defr target) (State, error) {
	switch s.awaiting.spellName {
	case "Pyrus Balio":
		return s.DoDmg(int(defr.pID), defr.id, PyrusBalioDmg), nil
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
	atkrCard := &s.field[atkr.pID][atkr.id]

	atk := atkrCard.atk0
	if atkr.atkNum == 1 {
		atk = atkrCard.atk1
	}

	if targetType == Permanent {
		switch atk.Name {
		case "removePerm":
			s.perms[defr.pID][defr.id] = Perm{} 
		}
		return s, errors.New("Unexpected attack")
	}


	defrCard := &s.field[defr.pID][defr.id]
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

		s = s.DoDmg(int(defr.pID), defr.id, atk.Dmg)

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

func (s State) awaitSpell(spell Instant) State {
	s.awaiting = Await{
		isTrue: true,
		spell: true,
		spellName: spell.name, 
	}
	return s
}

func (s State) cancelAwait() State {
	s.awaiting = Await{}
	return s
}
