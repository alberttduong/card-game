package game

import (
	"errors"
)

var TargetBububliusErr = GameErr{"Target protected by Bubublius"}
var TargetAreaErr = TargetErr{"Target's Area is invalid"}
var TargetPlayerErr = TargetErr{"Target's PlayerID is invalid"}
var TargetWizardErr = TargetErr{"Target Wizard doesn't exist"}
var TargetPermErr = TargetErr{"Perm Not Found"}
var TargetDeckErr = TargetErr{"Card Not Found Deck"}
var TargetDragonErr = ImplmtErr{"Target Dragon doesn't exist"}

type ImplmtErr struct {
	msg string
}

func (i ImplmtErr) Error() string {
	return i.msg
}

type GameErr struct {
	msg string
}

func (g GameErr) Error() string {
	return g.msg
}

type TargetErr struct {
	msg string
}

func (t TargetErr) Error() string {
	return t.msg
}

type DragoniusAtkErr struct{}

func (d DragoniusAtkErr) Error() string {
	return "Tried to attack with Dragonius"
}

type InputErr struct {
	msg string
}

func (i InputErr) Error() string {
	return i.msg
}

// Funcs

func (g State) checkPlayerID(p playerID) error {
	if n := int(p); n < 0 || n >= g.NumPlayers {
		return errors.New("Invalid Player ID")
	}
	return nil
}

func (g State) checkTarget(t target) error {
	if t.area == InstantSpell {
		return TargetAreaErr
	}

	if err := g.checkPlayerID(t.pID); err != nil {
		return TargetPlayerErr
	}

	switch t.area {
	case Wizard:
		if t.id < 0 || t.id >= len(g.Field[t.pID]) { 
			return TargetWizardErr
		}
	case Permanent:
		_, ok := g.Permanents[PermTarget{t.pID, t.id}]
		if !ok {
			return TargetPermErr
		}
	case Dragon:
		_, ok := g.Permanents[PermTarget{t.pID, t.id}]
		if !ok {
			return TargetDragonErr
		}
		empty := Card{}
		if g.Dragons[t.pID][t.id] == empty {
			return TargetDragonErr
		}
	default:
		return ImplmtErr{"Target area invalid"}
	}
	return nil
}

func (g State) inDeck(cname CardName) (int, bool) {
	for i, c := range g.Players[g.CurrentPlayer].deck {
		if c == cname {
			return i, true
		}
	}
	return 0, false
}
