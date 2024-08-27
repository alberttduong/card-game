package game

import (
	"fmt"
	"errors"
)

func (s State) startTurn() State {
	mana := s.players[s.currentPlayer].manaCap
	if mana < s.manaMax {
		s.players[s.currentPlayer].manaCap++
	}
	s.Mana = mana
	return s
}

func (s State) play(p playerID, c Card) (State, error) {
	if len(s.field[p]) == MaxFieldLen {
		return s, errors.New("Field is at max capacity")
	}
	s.field[p] = append(s.field[p], c)
	return s, nil
}

func (s State) drawCard(p playerID) (State, error) {
	deck := s.players[p].deck
	if len(deck) == 0 {
		return s, errors.New("Deck empty")
	}

	s.players[p].hand = append(s.players[p].hand, deck[len(deck)-1])
	s.players[p].deck = deck[:len(deck)-1]
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
	pID playerID
	id  int
}

// int(CardName)
func (s State) playSpell(spell CardName, defr target) State {
	switch spell {
		/*
	case fireball:
		s.field[defr.pID][defr.id].hp -= 1
		*/
	default:
		fmt.Println("no")
	}
	return s
}

func (s State) endTurn() (State, error) {
	s.currentPlayer = playerID(int(s.currentPlayer + 1) % s.numPlayers)
	return s.startTurn(), nil
}

func (s State) setMana(n int) State {
	s.Mana = n 
	return s
}

type Attacker struct {
	pID playerID
	id int
	usingAtk1 bool
}

func (s State) attack(atkr Attacker, defr target) State {
	atkrCard := s.field[atkr.pID][atkr.id]
	atk := atkrCard.atk1
	
	switch atk.Name {
	case "draw1":
		s, _ = s.drawCard(s.currentPlayer)
	}

	s.field[defr.pID][defr.id].hp -= atk.Dmg 
	return s
}
