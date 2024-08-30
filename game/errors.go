package game

import (
	"errors"
)

func (g State) checkPlayerID(p playerID) error {
	if n := int(p); n < 0 || n >= g.numPlayers {
		return errors.New("Invalid Player ID")
	}
	return nil
}

func (g State) checkTarget(t target) error {
	if t.area == InstantSpell {
		return errors.New("Invalid area")
	}

	if err := g.checkPlayerID(t.pID); err != nil {
		return errors.New("Target doesn't exist: Invalid Player")
	}

	var isValidID func (target) bool 

	switch t.area {
	case Wizard:
		isValidID = func (t target) bool {
			return t.id >= 0 && t.id < len(g.field[t.pID]) &&
					(t.atkNum == 0 || t.atkNum == 1)
		}
	case Permanent:
		isValidID = func (t target) bool {
			return true
		}
	}
	if isValidID(t) {
		return nil 
	}

	return errors.New("unexpected card type")
}
