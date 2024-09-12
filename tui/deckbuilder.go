package tui
import (
	"github.com/alberttduong/card-game/game"
	"github.com/nsf/termbox-go"
	"io"
	"os"
	"fmt"
	"slices"
)

const (
	DECK1_PATH = "deck1.txt"
	DECK2_PATH = "deck2.txt"
	Columns = 9
	DeckListWidth = 15
	DeckListHeight = 19 
	DefaultNumButtons = 4 
)

type DeckBuilder struct {
	editPath string
	Cards []game.Cdata
	cursor *Cursor
	numCards int

	errorMsg string
}

type Entries struct { deck []game.DeckEntry }

func NewDeckBuilder(cards []game.Cdata) *DeckBuilder {
	options := []Coord{} 
	n := len(cards) - 3
	for i := range n / Columns {
		options = append(options, Coord{i, Columns})
	}
	if remainder := n % Columns; remainder != 0 {
		options = append(options, Coord{n / Columns, remainder})
	}
	options = append(options, Coord{len(options), DefaultNumButtons})

	return &DeckBuilder{
		editPath: DECK1_PATH,
		Cards: cards, 
		numCards: n,
		cursor: &Cursor{
			Coords : options, 
		},
	}
}

func (s *DeckBuilder) Cursor() *Cursor {
	return s.cursor
}

func (s *DeckBuilder) Redraw() {
	clearScreen()

	middle := slices.Concat(
		s.CardGrid(),
		s.rowButtons(),
		[]string{
			s.errorMsg,
		},
	)
		
	middle = concatMany(
		s.DeckList(DECK1_PATH),
		s.DeckList(DECK2_PATH),
		CardPreview(s.CardViewContent()),
		middle,
	)

	scrn := slices.Concat(
		deckBuilderHeader(),
		middle,
	)
	render(scrn)
}

func (s DeckBuilder) rowButtons() []string {
	buttons := [][]string{
		Box("Edit Deck1"),
		Box("Edit Deck2"),
		Box("clear"),
		Box("exit"),
	}

	if s.cursor.Selected.y == len(s.cursor.Coords) - 1 {
		for i := range buttons {
			if s.cursor.Selected.x == i {
				buttons[i] = yellow(buttons[i])
			}
		}
	}

	if s.editPath == DECK1_PATH {
		buttons[0] = green(buttons[0])
	} else {
		buttons[1] = green(buttons[1])
	}
	
	return concatMany(buttons...) 
}

func (s DeckBuilder) CardViewContent() (data *TextWrapIter) {
	data = NewTextWrapIter(CardViewWidth)
	x, y := s.cursor.Selected.x, s.cursor.Selected.y 
	cell := y * Columns + x

	if cell >= s.numCards {
		return
	}

	card := s.Cards[cell]  
	data.AddLines(card.CName.String(), fmt.Sprintf("(%s)", card.Type), "")
	if card.Desc != "" {
		data.AddParagraph(card.Desc)
	}
	data.AddWizardAttackDesc(card)
	return
}

func (s *DeckBuilder) HandleEvent(ev termbox.Event) error {
	if ev.Ch == 'b' {
		return BACK
	}
	switch ev.Key {
	case termbox.KeyEnter:
		if s.cursor.Selected.y != len(s.cursor.Coords) - 1 {
			s.AddCard()
			break
		}

		switch s.cursor.Selected.x {
		case 0:
			s.editPath = DECK1_PATH
		case 1:
			s.editPath = DECK2_PATH
		case 2:
			s.ClearDeck()
		case 3:
			return BACK
		}
	case termbox.KeyBackspace2:
		if s.cursor.Selected.y != len(s.cursor.Coords) - 1 {
			s.RemoveCard()
		}
	}
	s.Redraw()
	return nil
}

func deckBuilderHeader() []string {
	text := make([]string, 4) 
	text[1] = fmt.Sprintf("%3s | %s", GameTitle, "Deck Builder")
	return text
}

func addRow(top, bottom []string) {
	top = slices.Concat(top, bottom)
}

func DeckMapFromFile(path string) (m map[int]int, e error) {
	deckFile, err := os.Open(path)
	if err != nil {
		_, err = os.Create(path)
		if err != nil {
			return m, err
		}
	}

	data := make([]byte, 100)
	_, err = deckFile.Read(data)
	if err != nil {
		return m, err
	}
	d, err := game.ParseDeck(data)
	if err != nil {
		return m, err
	}

	if err != nil {
		return m, err
	}
	return d, nil
}

func (s DeckBuilder) DeckList(path string) (res []string) {
	dMap, err := DeckMapFromFile(path)	
	entries := game.SortedDeckList(s.Cards, dMap)

	content := NewTextWrapIter(DeckListWidth)
	
	content.AddLines(path)
	invalid := game.ValidateDeck(s.Cards, dMap)
	if invalid != nil {
		content.AddParagraph(fmt.Sprintf("(Invalid: %s)", invalid.Error()))
		content.AddLines("")
	} else {
		content.AddLines("(Valid)", "")
	}

	if err != nil {
		if err == io.EOF {
			content.AddLines("Empty")
		} else {
			content.AddLines(err.Error())
		}
	} else {
		for _, entry := range entries { 	
			content.AddParagraph(fmt.Sprintf("%dx %s", 
				entry.Amount, game.CardName(entry.ID)))
		}
	}

	list := make([]string, DeckListHeight)
	list[0] = boxTop(DeckListWidth) 
	for i := range len(list) - 2 {
		line, _ := content.Next()
		list[i+1] = boxMiddle(DeckListWidth, line)
	}
	list[DeckListHeight - 1] = boxBottom(DeckListWidth) 

	if s.editPath == path {
		c := Green
		if invalid != nil {
			c = Yellow	
		}
		list[1] = color(c, list[1])
	}


	return list 
}

//todo
func (s *DeckBuilder) RemoveCard() error {
	dMap, err := DeckMapFromFile(s.editPath)	
	if err == game.DeckFormatErr {
		return err
	}
	entries := game.SortedDeckList(s.Cards, dMap)	
	cardID := s.cursor.Selected.y * Columns + s.cursor.Selected.x + 1
	index, found := game.SearchSortedEntries(entries, cardID)

	if !found {
		return nil
	}

	entries[index].Amount--
	entries = game.SortEntries(entries)

	d := game.EntriesToBytes(entries) 
	os.WriteFile(s.editPath, d, 0644)
	return nil
}

func (s *DeckBuilder) AddCard() error {
	dMap, err := DeckMapFromFile(s.editPath)	
	if err == game.DeckFormatErr {
		return err
	}
	entries := game.SortedDeckList(s.Cards, dMap)	
	cardID := s.cursor.Selected.y * Columns + s.cursor.Selected.x + 1
	index, found := game.SearchSortedEntries(entries, cardID)
	if found {
		entries[index].Amount++	
	} else {
		entries = append(entries, game.DeckEntry{ID: cardID, Amount: 1})
	}

	game.SortEntries(entries)
	d := game.EntriesToBytes(entries) 
	os.WriteFile(s.editPath, d, 0644)
	return nil
}

func (s DeckBuilder) ClearDeck() {
	os.WriteFile(s.editPath, []byte{}, 0644)
}

func (s DeckBuilder) CardGrid() []string {
	result := []string{}

	newRow := []string{}
	// i dont know why you have to subtract 3 but u do
	for i := range s.numCards { 
		newCard := cardNameImg(s.Cards, game.CardName(i + 1))
		if s.cursor.IsSelected(i % Columns, i / Columns) {
			newCard = yellow(newCard)
		}

		newRow = concat(newRow, newCard) 
		if i % Columns == Columns - 1 {
			result = slices.Concat(result, newRow)	
			newRow = []string{}
		}
	}
	result = slices.Concat(result, newRow)	
	
	return result
}
