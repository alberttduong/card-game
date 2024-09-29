package game

import "slices"

var castSpellOnTarget = map[CardName]func(State, target) (State, error){
	PyrusBalio:   castDamage(PyrusBalioDmg),
	DracusPyrio:  castDracusPyrio,
	Protectio:    castProtectio,
	Cancelio:     castCancelio,
	AngeliDustio: castAngeliDustio,
	Vitalius:     castVitalius,
	Retrievio:    castRetrievio,
	Extractio:    castExtractio,
	// Attachments
	Enhancius: castAttach,
	Mortius:   castAttach,
	Armorius:  castAttach,
	Bubublius: castAttach,
}

func (s State) cardToCast(defr target) (*Card, error) {
	c, err := s.cardFromTarget(defr)
	if err != nil {
		return c, err
	}
	if c.attached == Bubublius {
		return c, TargetBububliusErr
	}
	return c, nil
}

func castDamage(dmg int) func(State, target) (State, error) {
	return func(s State, defr target) (State, error) {
		c, err := s.cardToCast(defr)
		if err != nil {
			return s, err
		}

		s.Output.Printf("Doing %d damage to %s", dmg, c.CName)
		return s.DoDmgToCard(c, dmg), nil
	}
}

func castDracusPyrio(s State, defr target) (State, error) {
	s.Players[s.CurrentPlayer].Hand = nil
	s.Output.Printf("%s removed all cards from their hand", s.Players[s.CurrentPlayer])
	return castDamage(DracusPyrioDmg)(s, defr)
}

func castProtectio(s State, defr target) (State, error) {
	c, err := s.cardToCast(defr)
	if err != nil {
		return s, err
	}
	s.Output.Printf("%s is protected from damage this round", c.CName)
	c.protected = true
	return s, nil
}

func castRetrievio(s State, defr target) (State, error) {
	if err := s.checkTarget(defr); err != nil {
		return s, err
	}
	pt := PermTarget{pID: defr.pID, id: defr.id}

	p, ok := s.Permanents[pt]
	if !ok {
		return s, TargetPermErr
	}

	s.Output.Printf("Retrieving %s", p.CName)
	s.Players[s.CurrentPlayer].Hand = append(s.Players[s.CurrentPlayer].Hand, p.CName)
	return s.removePerm(pt)
}

func castExtractio(s State, defr target) (State, error) {
	id, ok := s.inDeck(CardName(defr.id))
	if !ok {
		return s, TargetDeckErr
	}
	p := &s.Players[s.CurrentPlayer]
	lastId := len(p.deck) - 1
	p.deck[lastId], p.deck[id] = p.deck[id], p.deck[lastId]
	s.Output.Printf("%s put %s back into their hand", p.Name, CardName(defr.id))
	return s.drawCard(s.CurrentPlayer)
}

func castAttach(s State, defr target) (State, error) {
	c, err := s.cardToCast(defr)
	if err != nil {
		return s, err
	}
	c.attached = s.awaiting.spellName

	p, ok := s.Permanents[s.awaiting.perm]
	if !ok {
		return s, TargetPermErr
	}
	p.AttachedTo = defr
	s.Permanents[s.awaiting.perm] = p
	s.Output.Printf("Attached %s to %s", p.CName, c.CName)

	return s, nil
}

func castVitalius(s State, defr target) (State, error) {
	s, err := castAttach(s, defr)
	if err != nil {
		return s, err
	}

	c, _ := s.cardToCast(defr)
	s.Output.Printf("%s's %s's HP increased by %d", 
		s.Players[s.CurrentPlayer], c.CName, AngeliDustioHeal)
	return s.doRawDmg(c, -2), nil
}

func castCancelio(s State, defr target) (State, error) {
	return s.removePerm(PermTarget{pID: defr.pID, id: defr.id})
}

func castAngeliDustio(s State, defr target) (State, error) {
	c, err := s.cardToCast(defr)
	if err != nil {
		return s, err
	}
	s.Output.Printf("Healed %s by %d", c.CName, AngeliDustioHeal)
	return s.DoDmgToCard(c, -AngeliDustioHeal), nil
}

var SpellsThatAwait = []CardName{
	PyrusBalio,
	Protectio,
	Cancelio,
	AngeliDustio,
	DracusPyrio,
	Retrievio,
	Extractio,
}

func (s State) playInstant(spell Instant) (State, error) {
	if spell.CName == DracusPyrio &&
		len(s.Players[s.CurrentPlayer].Hand) == 0 {
		return s, GameErr{"Can't play Dracus Pyrio with an empty Hand"}
	}

	s.Output.Printf("%s played %s", s.Players[s.CurrentPlayer].Name, spell.CName)
	if spell.CName == Extractio {
		if len(s.Players[s.CurrentPlayer].deck) == 0 {
			return s, GameErr{"Deck empty"}
		}
		s = s.printCardsInDeck()
	}

	if slices.Contains(SpellsThatAwait, spell.CName) {
		return s.awaitSpell(spell), nil
	}

	if spell.CName == Dralio {
		n := 0
		var err error	
		for range 2 {
			s, err = s.drawCard(s.CurrentPlayer)
			if err == nil {
				n++		
			}
		}

		s.Output.Printf("%s drew %d cards",
			s.Players[s.CurrentPlayer], n)

		if err != nil {
			s.Output.Printf("%s", err)
		}

		return s, nil 
	}

	return s, ImplmtErr{"Invalid instant"}
}
