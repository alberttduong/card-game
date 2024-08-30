package game

import (
	_ "embed"
	"fmt"
	"testing"
)

//go:embed cards.json
var data []byte
var cards []Cdata = GetCardData(data)

func (g State) doAtk(n int) State {
	if n != 0 && n != 1 {
		panic("atknum must be 0 or 1")
	}

	g, err := g.attack(target{pID: 0, id: 0, atkNum: n},
		target{pID: 0, id: 1})

	if err != nil {
		panic(err)
	}

	return g
}

func (g State) playCards(p int, names ...CardName) State {
	for _, n := range names {
		var err error
		g, err = g.play(playerID(p), CardFromName(cards, n))
		if err != nil {
			panic(err)
		}
	}	
	return g
}

func (g State) CardHp(p int, id int) int {
	return g.field[p][id].hp
}

func (g State) InitFullDeck() State {
	g.players[g.currentPlayer].deck = []CardName{}

	for i := 0; i < 30; i++ {
		g.players[g.currentPlayer].deck = append(g.players[g.currentPlayer].deck, Librarian)
	}

	return g
}

func Test_Mana(t *testing.T) {
	g, _ := InitState(2)
	expected := 1
	if g.Mana != expected {
		t.Errorf("Expected mana %d got %d", expected, g.Mana)
	}
	g, _ = g.endTurn()
	if g.Mana != expected {
		t.Errorf("Expected mana %d got %d", expected, g.Mana)
	}

	expected = 2
	g, _ = g.endTurn()
	if g.Mana != expected {
		t.Errorf("Expected mana %d got %d", expected, g.Mana)
	}
	g, _ = g.endTurn()
	if g.Mana != expected {
		t.Errorf("Expected mana %d got %d", expected, g.Mana)
	}
}

func Test_Librarian(t *testing.T) {
	g, _ := InitState(2)
	g = g.playCards(0, Librarian, Librarian)

	g = g.InitFullDeck()
	
	handSize := func () int {
		return len(g.players[0].hand)
	}
	hand := handSize() 
	g = g.doAtk(0)
	if newHand := handSize(); newHand != hand+1 {
		t.Errorf("Expected handsize: %d, Got: %d", hand+1, newHand)
		fmt.Println(g)
	}

	hp := g.CardHp(0, 1)

	for i := 0; i < 6; i++ {
		g, _ = g.drawCard(g.currentPlayer)
	}

	g = g.doAtk(1)
	if newHp := g.CardHp(0, 1); newHp != hp-3 {
		t.Errorf("Wrong HP expected %d got %d", hp-3, newHp)
		fmt.Println(g)
	}
}

func Test_Angel(t *testing.T) {
	g, _ := InitState(2)

	g = g.playCards(0, Angel, Librarian)

	g = g.DoDmg(0, 1, 3)
	g = g.doAtk(0)

	if hp := g.CardHp(0, 1); hp != 7 {
		t.Errorf("Expected: 7, Got: %d", hp)
	}

	g = g.DoDmg(0, 1, 8)
	g = g.doAtk(1)

	g, err := g.target(target{pID: 0, id: 1})
	if err != nil {
		t.Error(err)
	}

	if hp := g.CardHp(0, 1); hp != 8 {
		t.Errorf("Expected: 8, Got: %d", hp)
	}
	if hp := g.CardHp(0, 0); hp != 0 {
		t.Errorf("Expected: 0, Got: %d", hp)
	}
}

func Test_Magician(t *testing.T) {
	g, _ := InitState(2)

	g = g.playCards(0, Magician, Librarian)

	g = g.DoDmg(0, 0, 1)
	g = g.doAtk(0)
	if h := g.players[0].magicianHealth; h != 7 {
		t.Errorf("expected hp 7 got %d", h)
	}

	g, err := g.play(0, CardFromName(cards, g.players[0].hand[0]))
	if err != nil {
		t.Error(err)
	}

	if hp := g.CardHp(0, 0); hp != 7 {
		t.Errorf("expected hp 7 got %d", hp)
	}
	g = g.doAtk(1)
	if h := g.players[0].magicianHealth; h != 5 {
		t.Errorf("expected hp 5 got %d", h)
	}
}

func Test_Pyromancer(t *testing.T) {
	g, _ := InitState(2)
	g = g.playCards(0, Pyromancer, Librarian, Librarian)
	g = g.playCards(1, Librarian, Librarian)

	g = g.doAtk(0)
	for i := 0; i < 3; i++ {
		if hp := g.CardHp(0, i); hp != 7 {
			t.Errorf("expected hp 7 got %d", hp)
		}
	}

	g = g.doAtk(1)
	for p := 0; p < g.numPlayers; p++ {
		for i := 0; i < 2; i++ {
			if hp := g.CardHp(p, i); hp != 7 {
				t.Errorf("megasplash: expected hp 7 got %d", hp)
			}
		}
	}
}

func Test_Shieldmancer(t *testing.T) {
	g, _ := InitState(2)

	g = g.playCards(0, Shieldmancer, Librarian)

	g = g.doAtk(0)
	hp := g.CardHp(0, 1)
	g = g.DoDmg(0, 1, 8)
	if newHp := g.CardHp(0, 1); newHp != hp {
		t.Errorf("expected hp %d got %d", hp, newHp)
	}
	g, _ = g.endTurn()
	g, _ = g.endTurn()

	// protect only lasts 1 turn
	hp = g.CardHp(0, 1)
	g = g.DoDmg(0, 1, 3)
	if newHp := g.CardHp(0, 1); newHp != hp-3 {
		t.Errorf("expected hp %d got %d", hp, newHp)
	}

	g = g.doAtk(1)
	hp = g.CardHp(0, 1)
	g = g.DoDmg(0, 1, 2)
	if newHp := g.CardHp(0, 1); newHp != hp-1 {
		t.Errorf("expected hp %d got %d", hp, newHp)
	}
}

func Test_Conjurer(t *testing.T) {
	g, _ := InitState(2)

	g = g.playCards(0, Conjurer, Librarian)

	g0 := g
	mana := g0.Mana
	g0, _ = g0.doAtk(0).endTurn()
	g0, _ = g0.endTurn()
	if newMana := g0.Mana; newMana != mana+2 {
		t.Errorf("expected mana %d got %d", mana+2, newMana)
	}

	g1, _ := InitStateUsingMana(2)
	g1, _ = g1.play(playerID(0), CardFromName(cards, Conjurer))
	g1, _ = g1.play(playerID(0), CardFromName(cards, Librarian))
		
	//refactor when attacks cost mana
	g1 = g1.doAtk(1)
	g1 = g1.setMana(0)
	if g1.Mana != 0 {
		t.Errorf("expected 0 mana got %d", g1.Mana)
	}
	if g1.players[0].discountSpell == false {
		t.Errorf("discountspell shoudl be true")
	}
	g1, err := g1.play(playerID(0), CardFromName(cards, PyrusBalio))
	if err != nil {
		t.Error(err)
	}
	if g1.Mana != 0 {
		t.Errorf("expected 0 mana got %d", g1.Mana)
	}
}

func Test_Mortician(t *testing.T) {
	g, _ := InitState(2)

	g = g.playCards(0, Mortician, Librarian, Librarian)
	g = g.playCards(1, Librarian)

	g, _ = g.attack(target{pID: 0, id: 0, atkNum: 0},
		target{pID: 1, id: 0})

	if hp := g.CardHp(0, 1); hp != 7 {
		t.Errorf("expected hp %d got %d", 7, hp)
	}
	if hp := g.CardHp(0, 2); hp != 7 {
		t.Errorf("expected hp %d got %d", 7, hp)
	}

	g = g.DoDmg(0, 1, 8)
	g = g.DoDmg(0, 2, 8)

	hp := g.CardHp(0, 0)
	g, _ = g.attack(target{pID: 0, id: 0, atkNum: 1},
		target{pID: 0, id: 0})
	if newHp := g.CardHp(0, 0); newHp != hp-6 {
		t.Errorf("expected hp %d got %d", hp-6, newHp)
	}
}

func Test_MindMage(t *testing.T) {
	g, _ := InitState(2)
	g = g.playCards(0, MindMage, Shieldmancer)
	g, _ = g.attack(target{pID: 0, id: 1, atkNum: 0},
		target{pID: 0, id: 1})	

	hp := g.CardHp(0, 0) - 2

	g, _ = g.attack(target{pID: 0, id: 0, atkNum: 0},
		target{pID: 0, id: 0})		
	
	if newHp := g.CardHp(0, 0); newHp != hp {
		t.Errorf("expected hp %d got %d", hp, newHp)
	}

	g, _ = g.play(playerID(0), CardFromName(cards, Aquarius))
	if l := len(g.perms[0]); l != 1 {
		t.Errorf("expected perm len %d got %d", 1, l)
	}	
	g = g.doAtk(1)

	p := Perm{}
	if g.perms[0][0] == p {
		t.Error("Unexpectedly empty perm")
	}

	g, _ = g.target(target{pID: 0, id: 0, area: Permanent})

	if g.perms[0][0] != p {
		t.Error("didnt remove perm")
	}
}

func Test_Bloodeater(t *testing.T) {
	g, _ := InitState(2)
	g = g.playCards(0, Bloodeater, Librarian)

	g = g.doAtk(0)
	if newHp := g.CardHp(0, 1); newHp != 7 {
		t.Errorf("expected hp 7 got %d", newHp)
	}
	g, _ = g.target(target{pID: 0, id: 1})
	if newHp := g.CardHp(0, 1); newHp != 6 {
		t.Errorf("expected hp 6 got %d", newHp)
	}
	
	g = g.playCards(0, Librarian)
	g = g.doAtk(1)
	g, err := g.target(target{pID: 0, id:2})
	if err == nil {
		t.Errorf("expected an error but got none")
	}
	g = g.doAtk(1)
	g, err = g.target(target{pID: 0, id:2})
	if err != nil {
		t.Error(err)
	}
	if newHp := g.CardHp(0, 2); newHp != 4 {
		t.Errorf("expected hp 4 got %d", newHp)
	}
}

func Test_SpellsCostMana(t *testing.T) {
	g, _ := InitStateUsingMana(2)
	g = g.setMana(0)
	g, err := g.play(playerID(0), CardFromName(cards, PyrusBalio))
	if err == nil {
		t.Errorf("expected a notenoughmana error")
	}

	g = g.setMana(1)
	g, err = g.play(playerID(0), CardFromName(cards, PyrusBalio))
	if err != nil {
		t.Error(err)
	}
	if g.Mana != 0 {
		t.Errorf("expected 0 mana got %d", g.Mana)
	}
}

func Test_PyrusBalio(t *testing.T) {
	g, _ := InitState(2)
	g = g.playCards(0, Librarian, PyrusBalio)

	g, _ = g.target(target{pID: 0, id: 0})
	if newHp := g.CardHp(0, 0); newHp != 7 {
		t.Errorf("expected hp 7 got %d", newHp)
	}
}
/*
func Test_Aquarius(t *testing.T) {
	g, _ := InitState(2)
	g, e := g.play(playerID(0), CardFromName(cards, Aquarius))
	if e != nil {
		t.Error(e)
	}
}
*/
