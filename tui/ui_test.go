package tui 

import (
	"slices"
	"github.com/alberttduong/card-game/game"
	"testing"
	"fmt"
)

var (
	CARD = cardImg(game.Card{})
	YELLOW = yellow(cardImg(game.Card{}))
	EX_FIELD = slices.Concat(horzFieldLine(), concat(CARD, CARD))
	EX_FIELD_CARD = concat(EX_FIELD, CARD) 
)

func Test_MaxWidth(t *testing.T) {
	cases := map[int][]string {
		4: []string{
			"123",
			"1",
			"1235",
		},
		3: []string{
			"123",
			"1",
			"",
		},
		FieldHeaderWidth + 1: fieldHeader("Games"),
		FieldLength: EX_FIELD, 
		FieldLength + CardWidth: EX_FIELD_CARD, 
	}

	for e, c := range cases {
		if actW := maxWidth(c); actW != e {
			render(c)
			fmt.Println(len(c))
			t.Errorf("expected width %d got %d",
					 e, actW) 
		}
	}
}

func Test_RuneLen(t *testing.T) {
	if diff := runeLen(YELLOW[0]) - runeLen(CARD[0]);
	   diff != 0 {
		t.Errorf("expected diff 0 got %d (yellow - not)", diff)
	}
}

func Test_LeftPad(t *testing.T) {
	if diff := runeLen(leftPad(10, CARD[0])) - runeLen(leftPad(10, YELLOW[0]));
	   diff != 0 {
		t.Errorf("expected diff 0 got %d (card - yellow)", diff)
	}
}

func Test_WrapText(t *testing.T) {
	type tcase struct {
		text string
		w    int
	}

	cases := map[tcase][]string {
		tcase{"A cat boy farboy a a a a", 6}: []string{
			"A cat",
			"boy",
			"farboy",
			"a a a",
			"a",
		},
	}

	for c, e := range cases {
		if res := wrapText(c.text, c.w); !slices.Equal(e, res) {   
			t.Errorf("Expected %v got %v", e, res)
		}
	}
}
