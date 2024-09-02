package game

var castSpellOnTarget = map[string]func (State, target) (State, error) {
	"Pyrus Balio": castDamage(PyrusBalioDmg),
	"Dracus Pyrio": castDamage(DracusPyrioDmg),
	"Protectio": castProtectio,
	"Cancelio": castCancelio, 
	"Angeli Dustio": castAngeliDustio,
	"Vitalius": castVitalius,
	"Retrievio": castRetrievio,
	"Extractio": castExtractio,
	// Attachments
	"Enhancius": castAttach, 
	"Mortius": castAttach,
	"Armorius": castAttach,
	"Bubublius": castAttach,
}

func (s State) cardToCast(defr target) (*Card, error) {
	c, err := s.cardFromTarget(defr)		
	if err != nil {
		return c, err
	}
	if c.attached == "Bubublius" {
		return c, TargetBububliusErr 
	}
	return c, nil 
}

func castDamage(dmg int) func (State, target) (State, error) { 
	return func (s State, defr target) (State, error) {
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

	p, ok := s.permanents[pt]
	if !ok {
		return s, TargetPermErr 
	}

	s.players[s.currentPlayer].hand = append(s.players[s.currentPlayer].hand, p.cname)
	return s.removePerm(pt)
}

func castExtractio(s State, defr target) (State, error) {
	id, ok := s.inDeck(CardName(defr.id))
	if !ok {
		return s, TargetDeckErr 
	}
	p := &s.players[s.currentPlayer]
	lastId := len(p.deck) - 1
	p.deck[lastId], p.deck[id] = p.deck[id], p.deck[lastId]
	return s.drawCard(s.currentPlayer) 
}

func castAttach(s State, defr target) (State, error) {
	c, err := s.cardToCast(defr)	
	if err != nil {
		return s, err
	}
	c.attached = s.awaiting.spellName

	p, ok := s.permanents[s.awaiting.perm]
	if !ok {
		return s, TargetPermErr 
	}
	p.attachedTo = defr
	s.permanents[s.awaiting.perm] = p

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
