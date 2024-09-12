package game

import "slices"

var castSpellOnTarget = map[CardName]func(State, target) (State, error){
	PyrusBalio:   castDamage(PyrusBalioDmg),
	DracusPyrio:  castDamage(DracusPyrioDmg),
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

		return s.DoDmgToCard(c, dmg), nil
	}
}

func castProtectio(s State, defr target) (State, error) {
	c, err := s.cardToCast(defr)
	if err != nil {
		return s, err
	}
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

	return s, nil
}

func castVitalius(s State, defr target) (State, error) {
	s, err := castAttach(s, defr)
	if err != nil {
		return s, err
	}

	c, _ := s.cardToCast(defr)
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

	if slices.Contains(SpellsThatAwait, spell.CName) {
		return s.awaitSpell(spell), nil
	}

	if spell.CName == Dralio {
		s, _ = s.drawCard(s.CurrentPlayer)
		s, err := s.drawCard(s.CurrentPlayer)
		return s, err
	}

	return s, ImplmtErr{"Invalid instant"}
}
