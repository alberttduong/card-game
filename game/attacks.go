package game

import "slices"

var atkSideEffects = map[string]func(State, target, target) State{
	"moreMana":     moreMana,
	"draw1":        draw1,
	"protect":      protect,
	"reduce":       reduce,
	"nextDiscount": nextDiscount,
	"hurtAllies":   hurtAllies,
}

var SideEffectsAfterAtk = map[string]func(State, target, target) State{
	"disappear":    disappear,
	"disappear2":   disappear2,
	"splash":       splash,
	"megaSplash":   megaSplash,
	"removePerm":   withMsg("Preparing to remove a permanent", awaitAttacker),
	"revive":       withMsg("Preparing to fully heal an ally", awaitAttacker),
	"attackTwice":  withMsg("Preparing to attack again", awaitAttacker),
}

func moreMana(s State, atkr, defr target) State {
	s.Output.Printf("%s is conjuring more mana", s.Players[atkr.pID])
	s.Players[atkr.pID].moreMana++
	return s
}

func draw1(s State, atkr, defr target) State {
	return s.drawCards(s.CurrentPlayer, 1)
}

func awaitAttacker(s State, atkr, defr target) State {
	return s.setAwait(atkr)
}

func withMsg(msg string, f func (s State, t, t2 target) State) func (State, target, target) State {
	return func (s State, t, t2 target) State {
		s.Output.Printf(msg)
		return f(s, t, t2)
	}
}

func disappear(s State, atkr, defr target) State {
	s.Output.Printf("%s's Magician disappeared", s.Players[atkr.pID])
	atkrCard, _ := s.cardFromTarget(atkr)
	s.Players[s.CurrentPlayer].magicianHealth = atkrCard.HP
	h := s.Players[s.CurrentPlayer].Hand
	s.Players[s.CurrentPlayer].Hand = append(h, Magician)
	s.Field[atkr.pID] = slices.Delete(s.Field[atkr.pID], atkr.id, atkr.id+1) 
	return s
}

func disappear2(s State, atkr, defr target) State {
	s.Output.Printf("%s's Magician disappeared", s.Players[atkr.pID])
	s.Output.Printf("Magician might hurt itself")
	atkrCard, _ := s.cardFromTarget(atkr)
	newHp := atkrCard.HP - DisappearRecoil
	s.Players[s.CurrentPlayer].magicianHealth = newHp
	if newHp > 0 {
		h := s.Players[s.CurrentPlayer].Hand
		s.Players[s.CurrentPlayer].Hand = append(h, Magician)
		s.Field[atkr.pID] = slices.Delete(s.Field[atkr.pID], atkr.id, atkr.id+1) 
	} else {
		atkrCard.HP = 0
	}
	return s
}

func splash(s State, atkr, defr target) State {
	atkrCard, _ := s.cardFromTarget(atkr)
	dmg := s.baseDamage(atkr, atkrCard.Atk0)
	ally1, err := s.cardFromTarget(target{pID: defr.pID, id: (defr.id+1)%3}) 
	if err == nil {
		s = s.DoDmgToCard(ally1, dmg)
		s.Output.Printf("%s also attacked %s",
			atkrCard.CName,
			ally1.CName)
	}
	ally2, err := s.cardFromTarget(target{pID: defr.pID, id: (defr.id+2)%3}) 
	if err == nil {
		s = s.DoDmgToCard(ally2, dmg)
		s.Output.Printf("%s also attacked %s",
			atkrCard.CName,
			ally1.CName)
	}
	return s
}

func megaSplash(s State, atkr, defr target) State {
	for p := 0; p < s.NumPlayers; p++ {
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
	c, _ := s.cardFromTarget(atkr)
	s.Output.Printf("%s also attacked every other wizard",
			c.CName)
	return s
}

func protect(s State, atkr, defr target) State {
	s.Output.Printf("%s is protecting its allies", s.Field[atkr.pID][atkr.id].CName)
	protect := func(c *Card) {
		c.protected = true
	}

	s.applyToAllies(protect, atkr)
	return s
}

func reduce(s State, atkr, defr target) State {
	s.Output.Printf("%s's allies gained resistance", s.Field[atkr.pID][atkr.id].CName)
	setResistance := func(c *Card) {
		c.resistance = true
	}
	s.applyToAllies(setResistance, atkr)
	return s
}

func nextDiscount(s State, atkr, defr target) State {
	s.Output.Printf("Conjuring mana for the next spell")
	s.Players[atkr.pID].discountSpell = true
	return s
}

func hurtAllies(s State, atkr, defr target) State {
	hurt := func(c *Card) {
		s.DoDmgToCard(c, AllyRecoil)
		s.Output.Printf("Accidentally attacked %s's %s",
			s.Players[atkr.pID], c.CName.String())
	}
	s.applyToAllies(hurt, atkr)
	return s
}

func (s State) baseDamage(atkr target, atk Attack) int {
	atkrCard, err := s.cardFromTarget(atkr)
	if err != nil {
		panic(err)
	}

	if atkrCard.CName.String() == "Dragonius" {
		return DragoniusDmg
	}

	dmg := atk.Dmg
	if atkrCard.attached == Enhancius {
		dmg += EnhanciusBuff
	}
	switch atk.Name {
	case "dmgPerCard":
		// should be atkr id
		dmg += len(s.Players[s.CurrentPlayer].Hand) / CardPerDmg
	case "double":
		areDead := func(c *Card) bool {
			return c.HP < 1
		}
		if s.allAllies(areDead, atkr) {
			s.Output.Printf("%s's %s is using the blood of its allies to do double damage",
				s.Players[atkr.pID], atkrCard.CName)
			dmg *= 2
		}
	}
	return dmg
}
