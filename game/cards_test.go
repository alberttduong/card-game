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
	for p := 0; p < g.numPlayers; p++ {
		g.players[p].deck = []CardName{}
		for i := 0; i < 30; i++ {
			g.players[p].deck = append(g.players[p].deck, Librarian)
		}
	}

	return g
}

func (g State) checkAllyHpAt(t *testing.T, fID, expected int) {
	if hp := g.CardHp(0, fID); hp != expected {
		t.Errorf("Expected %d HP, Got %d", expected, hp)
	}
}

func (g State) checkHpIs(t *testing.T, expected int) {
	g.checkAllyHpAt(t, 0, expected)
}

func (g State) checkAllyHpIs(t *testing.T, expected int) {
	g.checkAllyHpAt(t, 1, expected)
}

func (g State) handSize() int {
	return len(g.players[g.currentPlayer].hand)
}

func (g State) checkHandSize(t *testing.T, expected int) {
	if size := g.handSize(); size != expected {
		t.Errorf("Expected hand size of %d, Got %d", expected, size)
		fmt.Println(g)
	}
}


func Test_LenPerm(t *testing.T) {
	cases := map[[MaxPermLen]Perm]int {
		[MaxPermLen]Perm{}: 0,
		[MaxPermLen]Perm{ Perm{} }: 0,
		[MaxPermLen]Perm{ Perm{name: "Perm"} }: 1,
		[MaxPermLen]Perm{ Perm{}, Perm{name: "Perm"} }: 1,
		[MaxPermLen]Perm{ Perm{name: "Perm"}, Perm{}, Perm{name: "Perm"} }: 2,
	}

	for c, e := range cases {
		if length := lenPerm(c); length != e {
			t.Errorf("Expected length %d, Got %d", e, length)
		}
	}
}

func Test_NumPerms(t *testing.T) {
	g, _ := NewTestGame(2)
	g = g.playCards(0, Dragonius, Dragonius, Librarian, Aquarius)
	g = g.playCards(1, Dragonius, Librarian) 

	type input struct {
		mineOnly bool
		name string
	}

	cases := map[input]int {
		input{false, "Dragonius"}: 3,
		input{false, "Aquarius"}: 1,
		input{false, "Librarian"}: 0,
		input{true, "Dragonius"}: 2,
		input{true, "Aquarius"}: 1,
		input{true, "Librarian"}: 0,
	}

	for n, e := range cases {	
		num := g.numOfPerms(n.name)		
		if n.mineOnly {
			num = g.numOfMyPerms(n.name)		
		}
		if num != e { 
			t.Errorf("Expected num perms %d, Got %d", e, num)
		}
	}
}

func Test_RemovePerm(t *testing.T) {
	// todo
}

func Test_Mana(t *testing.T) {
	g, _ := NewTestGame(2)
	expect := func (expected int) {
		if g.Mana != expected {
			t.Errorf("Expected mana %d got %d", expected, g.Mana)
		}
	}
	g, _ = g.endTurn()
	expect(1)

	g, _ = g.endTurn()
	expect(2)

	g, _ = g.endTurn()
	expect(2)
}

func Test_Librarian(t *testing.T) {
	g, _ := NewTestGame(2)
	g = g.playCards(0, Librarian, Librarian)

	g = g.InitFullDeck()
	
	hand := g.handSize()
	g = g.doAtk(0)
	g.checkHandSize(t, hand + 1)

	for i := 0; i < 6; i++ {
		g, _ = g.drawCard(g.currentPlayer)
	}

	hp := g.CardHp(0, 1)
	g = g.doAtk(1)
	g.checkAllyHpIs(t, hp - 3)
}

func Test_Angel(t *testing.T) {
	g, _ := NewTestGame(2)

	g = g.playCards(0, Angel, Librarian)

	g = g.DoDmg(0, 1, 3)
	g = g.doAtk(0)

	g.checkAllyHpIs(t, 7)

	g = g.DoDmg(0, 1, 8)
	g = g.doAtk(1)

	g, err := g.target(target{pID: 0, id: 1})
	if err != nil {
		t.Error(err)
	}

	g.checkAllyHpIs(t, 8)
	g.checkHpIs(t, 0)
}

func Test_Magician(t *testing.T) {
	g, _ := NewTestGame(2)

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

	g = g.doAtk(1)
	if h := g.players[0].magicianHealth; h != 5 {
		t.Errorf("expected hp 5 got %d", h)
	}
}

func Test_Pyromancer(t *testing.T) {
	g, _ := NewTestGame(2)
	g = g.playCards(0, Pyromancer, Librarian, Librarian)
	g = g.playCards(1, Librarian, Librarian)

	g = g.doAtk(0)
	for i := 0; i < 3; i++ {
		g.checkAllyHpAt(t, i, 7)
	}

	g = g.doAtk(1)
	for p := 0; p < g.numPlayers; p++ {
		for i := 0; i < 2; i++ {
			if hp := g.CardHp(p, i); hp != 7 {
				t.Errorf("Expected Hp 7, Got %d", hp)
			}
		}
	}
}

func Test_Shieldmancer(t *testing.T) {
	g, _ := NewTestGame(2)

	g = g.playCards(0, Shieldmancer, Librarian)

	g = g.doAtk(0)
	hp := g.CardHp(0, 1)
	g = g.DoDmg(0, 1, 8)
	g.checkAllyHpIs(t, hp)
	g, _ = g.endTurn()
	g, _ = g.endTurn()

	// protect only lasts 1 turn
	hp = g.CardHp(0, 1)
	g = g.DoDmg(0, 1, 3)
	g.checkAllyHpIs(t, hp - 3)

	g = g.doAtk(1)
	hp = g.CardHp(0, 1)
	g = g.DoDmg(0, 1, 2)
	g.checkAllyHpIs(t, hp - 1)
}

func Test_Conjurer(t *testing.T) {
	g, _ := NewTestGame(2)

	g = g.playCards(0, Conjurer, Librarian)

	g0 := g
	mana := g0.Mana
	g0, _ = g0.doAtk(0).endTurn()
	g0, _ = g0.endTurn()
	if newMana := g0.Mana; newMana != mana+2 {
		t.Errorf("expected mana %d got %d", mana+2, newMana)
	}

	g1, _ := NewGame(2)
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
	g, _ := NewTestGame(2)

	g = g.playCards(0, Mortician, Librarian, Librarian)
	g = g.playCards(1, Librarian)

	g, _ = g.attack(target{pID: 0, id: 0, atkNum: 0},
		target{pID: 1, id: 0})

	g.checkAllyHpAt(t, 1, 7)
	g.checkAllyHpAt(t, 2, 7)
	g = g.DoDmg(0, 1, 8)
	g = g.DoDmg(0, 2, 8)

	hp := g.CardHp(0, 0)
	g, _ = g.attack(target{pID: 0, id: 0, atkNum: 1},
		target{pID: 0, id: 0})
	g.checkHpIs(t, hp - 6)
}

func Test_MindMage(t *testing.T) {
	g, _ := NewTestGame(2)
	g = g.playCards(0, MindMage, Shieldmancer)
	g, _ = g.attack(target{pID: 0, id: 1, atkNum: 0},
		target{pID: 0, id: 1})	

	hp := g.CardHp(0, 0) - 2

	g, _ = g.attack(target{pID: 0, id: 0, atkNum: 0},
		target{pID: 0, id: 0})		
	
	g.checkHpIs(t, hp)

	g, _ = g.play(playerID(0), CardFromName(cards, Aquarius))
	g = g.doAtk(1)

	g, _ = g.target(target{pID: 0, id: 0, area: Permanent})

	if _, ok := g.permanents[PermTarget{0, 0}];
		  ok == true {
		t.Error("didnt remove perm")	
	}
}

func Test_Bloodeater(t *testing.T) {
	g, _ := NewTestGame(2)
	g = g.playCards(0, Bloodeater, Librarian)

	g = g.doAtk(0)
	g.checkAllyHpIs(t, 7)

	g, _ = g.target(target{pID: 0, id: 1})
	g.checkAllyHpIs(t, 6)
	
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
	g.checkAllyHpAt(t, 2, 4)
}

func Test_SpellsCostMana(t *testing.T) {
	g, _ := NewGame(2)
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
	g, _ := NewTestGame(2)
	g = g.playCards(0, Librarian, PyrusBalio)

	g, _ = g.target(target{pID: 0, id: 0})
	g.checkHpIs(t, 7)
}

func Test_Protectio(t *testing.T) {
	g, _ := NewTestGame(2)
	g = g.playCards(0, Librarian, Protectio)
	g, err := g.target(target{pID: 0, id: 0})
	if err != nil {
		t.Error(err)
	}	

	if g.field[0][0].protected == false {
		t.Error("expected wizard to be protected")
	}
}

func Test_Mortius(t *testing.T) {
	g, _ := NewTestGame(2)
	g = g.playCards(0, Magician, Librarian, Mortius)
	g, _ = g.target(target{pID: 0, id: 1})
	g = g.doAtk(1).doAtk(1)
	if g.field[0][1].attached != "Mortius" {
		t.Error("not attached")
	}
	g.checkHpIs(t, 6)
}

func Test_Enhancius(t *testing.T) {
	g, _ := NewTestGame(2)
	g = g.playCards(0, Librarian, Shieldmancer, Enhancius)
	g, _ = g.target(target{pID: 0, id: 0})
	g = g.doAtk(0)
	g.checkAllyHpIs(t, 5)
}

func Test_Dragonius(t *testing.T) {
	g, _ := NewTestGame(2)
	g = g.playCards(0, Librarian, Dragonius) 
	
	d, err := g.cardFromTarget(target{pID: 0, id: 0, area: Permanent})
	if err != nil {
		t.Error(err)
	}
	if d.name != "Dragonius" {
		fmt.Println(g.dragons)
		t.Errorf("Dragon card from target gives %v", d.name)
	}

	g, err = g.attack(target{pID: 0, id: 0, area: Permanent},
					target{pID: 0, id: 0})
	if err != nil {
		t.Error(err)
	}

	g, err = g.target(target{pID: 0, id: 0})
	if err != nil {
		t.Error(err)
	}
	if hp := g.field[0][0].hp; hp != 5 {
		t.Errorf("expected hp 5 got %d", hp)
	}
	
	if !g.permanents[PermTarget{0, 0}].activated {
		t.Error("expected activation of perm")
	}

	g, err = g.attack(target{pID: 0, id: 0},
			target{pID: 0, id: 0, area: Permanent})
	
	if err != nil {
		t.Error(err)
	}

	if d.hp != 2 {
		t.Errorf("Expected HP 2 Got %d", d.hp)
	}


	g, err = g.attack(target{pID: 0, id: 0, area: Permanent},
					target{pID: 0, id: 0})
	if err == nil {
		t.Error("expected an error")
	}

	g, _ = g.endTurn()
	if g.permanents[PermTarget{0, 0}].activated {
		t.Error("expected deactivation of perm")
	}
}

func Test_Cancelio(t *testing.T) {
	g, _ := NewTestGame(2)
	g = g.playCards(0, Dragonius, Cancelio)
	g, _ = g.target(target{pID: 0, id: 0, area: Permanent})
	if _, ok := g.permanents[PermTarget{0, 0}];
		  ok == true {
		fmt.Println(g.permanents)
		t.Error("expected empty perm")	
	}
}

func Test_Dralio(t *testing.T) {
	g, _ := NewTestGame(2)
	g = g.InitFullDeck()
	g = g.playCards(0, Dralio)
	g.checkHandSize(t, 2)
}

func Test_AngeliDustio(t *testing.T) {
	g, _ := NewTestGame(2)
	g = g.playCards(0, Librarian)
	g = g.DoDmg(0, 0, 1)
	g = g.playCards(0, AngeliDustio)
	g, _ = g.target(target{pID: 0, id: 0})
	g.checkHpIs(t, 8)
}

func Test_Conjorius(t *testing.T) {
	g, _ := NewTestGame(2) 
	g = g.playCards(0, Librarian, Librarian, Conjorius)
	g = g.DoDmg(0, 0, 8)
	g = g.DoDmg(0, 1, 8)
	g = g.DoDmg(0, 1, 8)
	if m := g.players[0].moreMana; m != 2 {
		t.Errorf("Expected more mana to be 2, got %d", m)
	}
}

func Test_Vitalius(t *testing.T) {
	g, _ := NewTestGame(2) 
	g = g.playCards(0, MindMage, Vitalius) 
	g, _ = g.target(target{pID: 0, id: 0})
	g.checkHpIs(t, 10)
	g, _ = g.attack(target{pID: 0, id: 0, atkNum: 1},
	 			    target{pID: 0, id: 0}) 
	g, _ = g.target(target{pID: 0, id: 0, area: Permanent})
	g.checkHpIs(t, 5)
}

func Test_Librarius(t *testing.T) {
	g, _ := NewTestGame(2) 
	g = g.InitFullDeck()

	gWithLib, _ := NewTestGame(2)
	gWithLib = gWithLib.InitFullDeck()
	gWithLib = gWithLib.playCards(0, Librarius)

	if g.handSize() != gWithLib.handSize() {
		t.Error("Expected equal handsizes")
	}
	g, _ = g.endTurn()	
	gWithLib, _ = gWithLib.endTurn()	

	if gWithLib.handSize() != g.handSize() + 1 {
		t.Errorf("Expected handsize to differ by 1, got %d and %d", gWithLib.handSize(), g.handSize() + 1)
	}
}

func Test_Aquarius(t *testing.T) {
	g, _ := NewTestGame(2)
	g = g.playCards(0, Aquarius)
	for i := 0; i < 15; i++ {
		g, _ = g.endTurn()
	}
	if g.Mana != 7 {
		t.Errorf("expected mana 7 got %d", g.Mana)
	}
}

func Test_Bubublius(t *testing.T) {
	g, _ := NewTestGame(2)
	g = g.playCards(0, Librarian, Bubublius)
	g, e := g.target(target{pID: 0, id: 0})
	if e != nil {
		t.Error(e)
	}
	g = g.playCards(0, PyrusBalio)
	g, err := g.target(target{pID: 0, id: 0})
	if err == nil {
		fmt.Println(g.field[0][0].attached)
		t.Error("expected an error got none")
	}
}

func Test_Meteorus(t *testing.T) {
	g, _ := NewTestGame(2)
	g = g.playCards(0, Librarian, Meteorus)
	g, _ = g.activatePerm(PermTarget{0, 0})
	g.checkHpIs(t, 7)
	g, _ = g.activatePerm(PermTarget{0, 0})
	g.checkHpIs(t, 7)
}

func Test_Armorius(t *testing.T) {
	g, _ := NewTestGame(2)
	g = g.playCards(0, Librarian, Armorius)
	g, _ = g.target(target{pID: 0, id: 0})	
	g = g.DoDmg(0, 0, 1)
	g.checkHpIs(t, 8)
	g = g.DoDmg(0, 0, 3)
	g.checkHpIs(t, 6)
}

func Test_DracusPyrio(t *testing.T) {
	g, _ := NewTestGame(2)
	g = g.InitFullDeck()
	g = g.playCards(0, Librarian)
	g, err := g.play(0, Instant{name: "DracusPyrio"})
	if err == nil {
		t.Error("expected error got none")
	}

	g, _ = g.drawCard(0)
	g, _ = g.play(0, Instant{name: "Dracus Pyrio"})
	g, _ = g.target(target{pID: 0, id: 0}) 
	g.checkHpIs(t, 1)
}

func Test_Retrievio(t *testing.T) {
	g, _ := NewTestGame(2)
	g = g.playCards(0, Aquarius, Retrievio)
	g, err := g.target(target{pID: 0, id: 0, area: Permanent})
	if err != nil {
		t.Error(err)
	}
	if g.handSize() != 1 {
		t.Errorf("Expected hand size of 1, Got %d", g.handSize())
	} else {
		if c := g.players[0].hand[0]; c != Aquarius {
			t.Errorf("Expected %v in hand, got %v",
					 Aquarius, c) 
		}
	}
	_, ok := g.permanents[PermTarget{0, 0}]
	if ok {
		t.Error("Permanent not removed")	
	}
}

func Test_Extractio(t *testing.T) {
	g, _ := NewTestGame(2)
	g = g.InitFullDeck()
	g = g.playCards(0, Extractio)
	g, err := g.target(target{id: 0, area: Deck})
	if err != nil {
		t.Error(err)
	}
	if g.players[0].hand[0] != Librarian {
		t.Error("Librarian not in hand")
	}

	g = g.playCards(0, Extractio)
	g, err = g.target(target{id: int(Angel), area: Deck})
	if err == nil {
		t.Error("expected err")
	}
	g.checkHandSize(t, 1)
}
