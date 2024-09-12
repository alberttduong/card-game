package game

import (
	"errors"
	"fmt"
	"math/rand"
	"slices"
)

const (
	CardsDrawnAtStart= 5
	CardPerDmg       = 2
	MaxHp            = 8
	DisappearRecoil  = 2
	AllyRecoil       = 1
	PyrusBalioDmg    = 1
	EnhanciusBuff    = 2
	DragoniusDmg     = 3
	AngeliDustioHeal = 2
	VitaliusBuff     = 2
	MeteorusDmg      = 1
	DracusPyrioDmg   = 7
)

func (s State) Start(cards []Cdata) State {
	//todo
	for p := range s.NumPlayers {
		slices.SortFunc(s.Players[p].deck, func (a, b CardName) int {
			return int(a) - int(b)
		})
		for range MaxWizards {
			s, _ = s.play(playerID(p), CardFromName(cards, s.Players[p].deck[0]))
			s.Players[p].deck = slices.Delete(s.Players[p].deck, 0, 1)
		}

		rand.Shuffle(len(s.Players[p].deck), func(i, j int) {
			s.Players[p].deck[i], s.Players[p].deck[j] = 
			s.Players[p].deck[j], s.Players[p].deck[i]
		})

		s = s.drawCards(playerID(p), CardsDrawnAtStart)
	}
	return s.startTurn()
}

func (s State) startTurn() State {
	p := &s.Players[s.CurrentPlayer]

	mana := p.manaCap
	manaMax := s.manaMax + s.numOfPerms(Aquarius)

	if mana < manaMax {
		p.manaCap++
	}

	mana += p.moreMana
	p.moreMana = 0

	s.Mana = mana

	field := &s.Field[s.CurrentPlayer]
	for i := 0; i < len(*field); i++ {
		(*field)[i].protected = false
	}

	s.output.Println(fmt.Sprintf("Player %v's turn", s.CurrentPlayer))
	if !s.testing {
		s = s.drawCards(p.ID, 1)
	}

	if numLib := s.numOfPerms(Librarius); numLib > 0 {
		s.output.Println("Librarius activated")
		s = s.drawCards(p.ID, numLib)
	}

	return s
}

func CardFromName(cards []Cdata, n CardName) Playable {
	c := cards[int(n)-1]

	if c.Type == "perm" {
		return Perm{
			Cost:  c.Hp,
			CName: c.CName,
		}
	}

	if c.Type == "instant" {
		return Instant{
			Cost:  c.Hp, // should be c.Cost
			CName: c.CName,
		}
	}

	return Card{
		HP:    c.Hp,
		Atk0:  c.Atk0,
		Atk1:  c.Atk1,
		CName: c.CName,
	}
}

func (s State) playPerm(player playerID, p Perm) (State, error) {
	s, pTarget, err := s.addPerm(player, p)
	if err != nil {
		return s, err
	}

	switch p.CName {
	case Mortius, Enhancius, Vitalius, Bubublius, Armorius: // Attachments
		s = s.awaitPerm(pTarget)
	case Dragonius:
		c := Card{
			CName: Dragonius,
			HP:    3,
			Atk0: Attack{
				Name: "Dragon Breath",
				Dmg:  3,
			},
		}
		s.Dragons[player][s.lenPermsOf(player)-1] = c
	case Aquarius, Conjorius, Librarius, Meteorus:
	default:
		return s, errors.New("unexpected perm")
	}
	return s, nil
}

func (s State) play(p playerID, c Playable) (State, error) {
	cardType := c.getType()

	if cardType == Wizard {
		if len(s.Field[p]) == MaxFieldLen {
			return s, errors.New("Field is at max capacity")
		}
		s.Field[p] = append(s.Field[p], c.(Card))
		return s, nil
	}

	if s.testing == false {
		cost := c.getCost()
		if s.Players[p].discountSpell {
			cost--
			s.Players[p].discountSpell = false
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

func (s State) drawCards(p playerID, n int) State {
	drawn := 0
	for range n {
		newS, err := s.drawCard(p)
		if err == nil {
			drawn++
		}
		s = newS
	}
	s.output.Printf("Player %d drew %d cards\n", p, drawn)
	return s
}

func (s State) drawCard(p playerID) (State, error) {
	player := &s.Players[p]
	d := player.deck
	if len(d) == 0 {
		//s.output.Println("Deck empty")
		return s, errors.New("Deck empty")
	}
	player.Hand = append(player.Hand, d[len(d)-1])
	player.deck = d[:len(d)-1]
	return s, nil
}

func (s State) removeFromHand(p playerID, idx int) (State, error) {
	hand := s.Players[p].Hand
	if idx >= len(hand) || idx < 0 {
		return s, errors.New("Index out of bounds")
	}
	hand[idx] = hand[len(hand)-1]
	hand = hand[:len(hand)-1]
	s.Players[p].Hand = hand
	return s, nil
}

type target struct {
	pID    playerID
	area   cardType
	id     int
	atkNum int
}

func (s State) endTurn() (State, error) {
	s.CurrentPlayer = playerID(int(s.CurrentPlayer+1) % s.NumPlayers)
	for k, v := range s.Permanents {
		v.Activated = false
		s.Permanents[k] = v
	}
	return s.startTurn(), nil
}

func (s State) setMana(n int) State {
	s.Mana = n
	return s
}

func (s State) activatePerm(pt PermTarget) (State, error) {
	p, ok := s.Permanents[pt]
	if !ok {
		return s, errors.New("couldnt find perm")
	}

	if p.Activated {
		return s, errors.New("Perm alr activated")
	}
	p.Activated = true
	s.Permanents[pt] = p

	if p.CName == Meteorus {
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
		d := PermTarget{atkr.pID, atkr.id}
		dragon, ok := s.Permanents[d]
		if !ok {
			return s, errors.New("couldnt find dragon")
		}
		if dragon.Activated {
			return s, errors.New("attack with dragon only once per turn")
		}
		dragon.Activated = true
		s.Permanents[d] = dragon
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

	defrWasAlive := defrCard.HP > 0

	s = s.DoDmgToCard(defrCard, dmg)

	if defrWasAlive && defrCard.HP == 0 {
		if atk.Name == "frenzy" {
			s = s.setAwait(atkr)
		}
		if defrCard.attached == Mortius {
			s = s.DoDmgToCard(atkrCard, 2)
		}
	}

	return s, nil
}

func (s State) allAllies(f func(*Card) bool, t target) bool {
	for i := 0; i < MaxFieldLen-1; i++ {
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
	for i := 0; i < MaxFieldLen-1; i++ {
		fId := (t.id + 1 + i) % MaxFieldLen
		card, err := s.cardFromTarget(target{pID: t.pID, id: fId})
		if err != nil {
			continue
		}
		f(card)
	}
}

func (g State) DoDmg(p, id, dmg int) State {
	return g.DoDmgToCard(&g.Field[p][id], dmg)
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
			if c.attached == Armorius {
				dmg--
			}
		}
	}

	return g.doRawDmg(c, dmg)
}

func (g State) doRawDmg(c *Card, dmg int) State {
	maxH := MaxHp
	if c.attached == Vitalius {
		maxH += VitaliusBuff
	}

	newHp := c.HP - dmg
	if newHp < 0 {
		newHp = 0
	} else if newHp > maxH {
		newHp = maxH
	}

	if c.HP > 0 && newHp == 0 {
		for i, _ := range g.Players {
			g.Players[i].moreMana += g.numOfPermsOf(playerID(i), Conjorius)
		}
	}
	c.HP = newHp

	return g
}

func (s State) expectedTargetType() cardType {
	atkr := s.awaiting.atkr
	atkrCard := &s.Field[atkr.pID][atkr.id]

	name := atkrCard.Atk0.Name
	if atkr.atkNum == 1 {
		name = atkrCard.Atk1.Name
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
		p := s.Permanents[PermTarget{t.pID, t.id}]
		if p.CName == Dragonius {
			return &s.Dragons[t.pID][t.id], nil
		}
		return nil, errors.New("perm is not a valid target")
	}

	if t.area != Wizard {
		return nil, errors.New("Unexpected area in target")
	}
	return &s.Field[t.pID][t.id], nil
}

func (s State) spellTarget(defr target) (State, error) {
	spell, ok := castSpellOnTarget[s.awaiting.spellName]
	if !ok {
		return s, errors.New("Invalid spell name")
	}

	return func() (State, error) {
		s, e := spell(s, defr)
		s = s.cancelAwait()
		return s, e
	}()
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
			return s.removePerm(PermTarget{defr.pID, defr.id})
		}
		return s, errors.New("Unexpected attack")
	}

	switch atk.Name {
	case "revive":
		atkrCard.HP = 0
		defrCard.HP = MaxHp
	case "attackTwice":
		s, err := s.attack(atkr, defr)
		s = s.cancelAwait()
		return s, err
	case "frenzy":
		defrWasAlive := defrCard.HP > 0

		s, err := s.attack(atkr, defr)
		if err != nil {
			return s, nil
		}

		if defrWasAlive && defrCard.HP == 0 {
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
		isTrue:    true,
		spell:     true,
		spellName: spell.getCardName(),
	}
	return s
}

func (s State) AwaitStatus() string {
	return fmt.Sprintf("%t %s", s.awaiting.isTrue, s.awaiting.spellName.String()) 
}

func (s State) awaitPerm(pt PermTarget) State {
	p, ok := s.Permanents[pt]
	if !ok {
		panic("COuldnt find perm to await")
	}
	s.awaiting = Await{
		isTrue:    true,
		spell:     true,
		spellName: p.getCardName(),
		perm:      pt,
	}
	return s
}

func (s State) cancelAwait() State {
	s.awaiting = Await{}
	return s
}

func (s State) randomTarget() (target, error) {
	targets := []target{}
	for p := range s.Players {
		for i := range MaxFieldLen {
			t := target{pID: playerID(p), id: i}
			c, err := s.cardFromTarget(t)
			if err != nil {
				continue
			}

			if c.HP != 0 {
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
	if c.CName == Dragonius {
		return Attack{}, DragoniusAtkErr{}
	}

	if n != 0 && n != 1 {
		panic("invalid atk number")
	}
	if n == 1 {
		return c.Atk1, nil
	}
	return c.Atk0, nil
}

type PublicPermID struct {
	PID, ID int
}

// In order of players
func (s State) SortedPerms() [][]PublicPermID {
	keys := make([][]PublicPermID, s.NumPlayers)

	for k, _ := range s.Permanents {
		keys[k.pID] = append(keys[k.pID],
			PublicPermID{int(k.pID), k.id})
	}
	for i, _ := range keys {
		slices.SortStableFunc(keys[i], func(a, b PublicPermID) int {
			return a.ID - b.ID
		})
	}
	return keys
}

func (s State) GetPerm(p PublicPermID) (Perm, bool) {
	c, b := s.Permanents[PermTarget{playerID(p.PID), p.ID}]
	return c, b
}
