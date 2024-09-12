package game

var atkSideEffects = map[string]func(State, target, target) State{
	"moreMana":     moreMana,
	"draw1":        draw1,
	"disappear":    disappear,
	"disappear2":   disappear2,
	"splash":       splash,
	"megaSplash":   megaSplash,
	"protect":      protect,
	"reduce":       reduce,
	"nextDiscount": nextDiscount,
	"hurtAllies":   hurtAllies,
	"revive":       awaitAttacker,
	"attackTwice":  awaitAttacker,
	"removePerm":   awaitAttacker,
}

func moreMana(s State, atkr, defr target) State {
	s.Players[atkr.pID].moreMana++
	return s
}

func draw1(s State, atkr, defr target) State {
	s, _ = s.drawCard(s.CurrentPlayer)
	return s
}

func awaitAttacker(s State, atkr, defr target) State {
	return s.setAwait(atkr)
}

func disappear(s State, atkr, defr target) State {
	atkrCard, _ := s.cardFromTarget(atkr)
	s.Players[s.CurrentPlayer].magicianHealth = atkrCard.HP
	h := s.Players[s.CurrentPlayer].Hand
	s.Players[s.CurrentPlayer].Hand = append(h, Magician)
	return s
}

func disappear2(s State, atkr, defr target) State {
	atkrCard, _ := s.cardFromTarget(atkr)
	newHp := atkrCard.HP - DisappearRecoil
	s.Players[s.CurrentPlayer].magicianHealth = newHp
	if newHp > 0 {
		h := s.Players[s.CurrentPlayer].Hand
		s.Players[s.CurrentPlayer].Hand = append(h, Magician)
	}
	return s
}

func splash(s State, atkr, defr target) State {
	atkrCard, _ := s.cardFromTarget(atkr)
	dmg := s.baseDamage(atkr, atkrCard.Atk0)
	s = s.DoDmg(int(defr.pID), (defr.id+1)%3, dmg)
	s = s.DoDmg(int(defr.pID), (defr.id+2)%3, dmg)
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
	return s
}

func protect(s State, atkr, defr target) State {
	protect := func(c *Card) {
		c.protected = true
	}

	s.applyToAllies(protect, atkr)
	return s
}

func reduce(s State, atkr, defr target) State {
	setResistance := func(c *Card) {
		c.resistance = true
	}
	s.applyToAllies(setResistance, atkr)
	return s
}

func nextDiscount(s State, atkr, defr target) State {
	s.Players[atkr.pID].discountSpell = true
	return s
}

func hurtAllies(s State, atkr, defr target) State {
	hurt := func(c *Card) {
		s.DoDmgToCard(c, AllyRecoil)
	}
	s.applyToAllies(hurt, atkr)
	return s
}

func (s State) baseDamage(atkr target, atk Attack) int {
	atkrCard, err := s.cardFromTarget(atkr)
	if err != nil {
		panic(err)
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
			dmg *= 2
		}
	}
	return dmg
}
