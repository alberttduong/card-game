package tui 

import (
	"github.com/alberttduong/card-game/game"
	"github.com/nsf/termbox-go"
	"strings"
	"slices"
	"fmt"
)

type area int
const (
	CardViewWidth = 22 
	CardViewHeight = 19

	Wizard area = iota
	Permanent
	Hand
)

var (
	EXIT = ScreenErr{"Exit"}
	BACK = ScreenErr{"Back"}
	UNKNOWN = ScreenErr{"Unknown key"}
)

type Coord struct {
	realRow, length int
}

type CardPos struct {
	x, y int
}

func (s Screen) Cursor() *Cursor {
	return s.cursor
}

type Screen struct {
	cursor *Cursor
	selected CardPos 

	command string
	Inp Input
	Output Input

	Cards []game.Cdata
	Game game.State
	Perms [][]game.PublicPermID
}

// TODO
func (s Screen) InitGame() (game.State, error) {
	g, err := game.NewGame(2)
	if err != nil {
		return g, err
	}
	
	deck1, err := DeckMapFromFile(DECK1_PATH)
	if err != nil {
		return g, err
	}
	deck2, err := DeckMapFromFile(DECK2_PATH)
	if err != nil {
		return g, err
	}
	g, err = g.SetDeckFromMap(0, s.Cards, deck1) 
	if err != nil {
		return g, err
	}
	g, err = g.SetDeckFromMap(1, s.Cards, deck2) 
	if err != nil {
		return g, err
	}
	return g.Start(s.Cards), nil
}

func NewScreen(cards []game.Cdata, game game.State) *Screen {
	s := Screen{
		Cards: cards,
		cursor: &Cursor{},
		Output: Input{leftAlign: true},
		selected: CardPos{-1, -1},
	}
	return &s
}

func (s *Screen) HandleEvent(ev termbox.Event) error {
	key := ev.Ch
	if !s.Inp.Active {
		switch key {
		case 'b':
			return BACK
		case 'i':
			s.Inp.Active = true
		case 'n':
			s.command = fmt.Sprintf("attack %s 0 ", s.cursor.String())
			s.selected = s.cursor.Selected
		case 'm':
			s.command = fmt.Sprintf("attack %s 1 ", s.cursor.String())
		case 'p':
			s.Inp.Reset()
			s.Execute("end")
		}

		switch ev.Key {
		case termbox.KeyEnter: 
			fmt.Println(s.cursor.Selected)
			if s.cursor.SelectedYis(s.Game.NumPlayers) {
 				cmd := fmt.Sprintf("play %d", s.cursor.Selected.x)
				s.Execute(cmd)
				break
			}

			if strings.HasPrefix(s.command, "attack") {
				s.Execute(s.command + s.cursor.String())
				s.command = ""
				s.resetSelected()
				break
			}

 			s.Execute(fmt.Sprintf("target %s", s.cursor.String()))
		}
		

		s.Redraw()
		return nil
	}

	// Keyboard
	switch ev.Key {
	case termbox.KeyCtrlQ: 
		s.Inp.Reset()
	case termbox.KeyBackspace2: 
		s.Inp.Backspace()	
	case termbox.KeyEnter:
		cmd := s.Inp.Reset()
		s.Execute(cmd)
	case termbox.KeySpace:
		s.Inp.Space()
	default:
		s.Inp.AddKey(key)
	} 
	s.Redraw()
	return nil
}

type ScreenErr struct { msg string }
func (e ScreenErr) Error() string { return e.msg }

func (s *Screen) Execute(command string) error {
	command = strings.Trim(command , "\n")
	args := strings.Split(command , " ")
	newG, err := s.Game.Execute(s.Cards, args...)
	s.Game = newG
	if err != nil {
		s.Output.Reset()
		s.Output.text = err.Error()
		return err
	}
	s.cursor.ResetCursor()
	return err
}

func (s *Screen) Update() {
	s.Perms = s.Game.SortedPerms()
	options := []Coord{}
	for i := range s.Game.NumPlayers {
		length := len(s.Game.Field[i]) + len(s.Perms[i])
		if length > 0 {
			options = append(options, Coord{realRow: i, length: length}) 
		}
	}
	length := len(s.Game.Players[s.Game.CurrentPlayer].Hand)
	if length > 0 {
		options = append(options, Coord{realRow: s.Game.NumPlayers, length: length}) 
	}
	s.cursor.Coords = options
}

func (s Screen) SelectedCardData() game.Cdata {
	//TODO
	hand := s.Game.Players[s.Game.CurrentPlayer].Hand
	if s.cursor.Selected.y == len(s.cursor.Coords) - 1 && len(hand) > 0 {
		return s.Cards[hand[s.cursor.Selected.x] - 1]
	}
	return s.Cards[6]
}



func (s Screen) CardViewContent() (data *TextWrapIter) {
	data = NewTextWrapIter(CardViewWidth)
	x, y := s.cursor.Selected.x, s.cursor.Selected.y 
	options := s.cursor.Coords	

	if y >= len(options) {
		return 
	}

	hand := s.Game.Players[s.Game.CurrentPlayer].Hand
	if y == len(options) - 1 && len(hand) > 0 {
		card := s.Cards[hand[x] - 1]  
		data.AddLines(card.CName.String(), "(In Hand)", "")
		data.AddParagraph(card.Desc)
		data.AddWizardAttackDesc(card)
		return
	}

	coord := options[y]
	lenField := len(s.Game.Field[coord.realRow])
	if x >= lenField {
		pt := s.Perms[coord.realRow][x - lenField] 
		perm, ok := s.Game.GetPerm(pt) 
		if !ok {
			panic("Perm not found")
		}

		data.AddLines(perm.CName.String(), "(Permanent)", "")
		cardData := s.Cards[int(perm.CName) - 1]
		data.AddParagraph(cardData.Desc)
		return
	}

	c := s.Game.Field[coord.realRow][x]
	data.AddLines(c.CName.String(), "(Wizard)", "") 
	data.AddLines(fmt.Sprintf("%d/%d ", c.HP, 8), "") 
	data.AddWizardAttackDesc(s.Cards[int(c.CName)-1])
	return
}

func view(title string, content *TextWrapIter) []string {
	view := make([]string, CardViewHeight)
	for i := range CardViewHeight - 3 {
		line, err := content.Next()
		if err != nil {
			line = ""
		}
		view[3+i] = boxLeftJustifyMiddle(CardViewWidth, line) 
	}

	view[0] = boxTop(CardViewWidth) 
	view[1] = boxLeftJustifyMiddle(CardViewWidth, title)
	view[2] = boxMiddleLine(CardViewWidth)
	view[len(view)-1] = boxBottom(CardViewWidth)
	return view
}

func (s Screen) ChatView() []string {
	content := NewTextWrapIter(CardViewWidth) 
	lines := strings.Split(s.Game.Logs.String(), "\n")
	for _, str := range lines {
		content.AddParagraph(str)
	}

	if l, chatL := len(content.lines), CardViewHeight-4;
	   l > chatL {
		content.lines = content.lines[l - chatL:]
	}

	return view("Game Log", content)
}

func CardPreview(content *TextWrapIter) []string {
	return view("Card Info", content) 
}

func CardView(s Screener, content *TextWrapIter) []string {
	view := make([]string, CardViewHeight)
	
	for i := range CardViewHeight - 2 {
		line, err := content.Next()
		if err != nil {
			line = ""
		}
		view[i+1] = boxLeftJustifyMiddle(CardViewWidth, line) 
	}
	view[0] = boxTop(CardViewWidth) 
	view[len(view)-1] = boxBottom(CardViewWidth)
	return view
}

func (s *Screen) Redraw() {
	clearScreen()
	s.Update()

	var scrn []string
	add := func(s []string) {
		scrn = slices.Concat(scrn, s)
	}

	add(fieldLineWithTopCross())	
	for p := range s.Game.NumPlayers {
		fieldPerm := s.field(p)
		add(fieldPerm)
		add(fieldLineWithCross())	
	}
	add(s.hand())
	add(fieldLineWithBottomCross())	
	
	scrn = concatMany(s.ChatView(), scrn, CardPreview(s.CardViewContent()))
	scrn = slices.Concat(s.gameHeader(), scrn,
		concat(s.Inp.Textbox(), s.Output.Textbox()),
		[]string{
			fmt.Sprintf("%v", s.cursor.Selected),
		},
	)
	render(scrn)
}

func (s Screen) hand() []string {
	text := fieldHeader("Your Hand", fmt.Sprintf("[Player%d]", s.Game.CurrentPlayer))
	for i, r := range s.Game.Players[s.Game.CurrentPlayer].Hand {
		rightText := cardNameImg(s.Cards, r)
		if s.cursor.IsSelected(i, s.Game.NumPlayers) {
			yellow(rightText) 
		}
		text = concat(text, rightText)
	}
	RightBorder(text)
	return text
}

func (s Screen) field(p int) []string {
	text := fieldHeader("Field", fmt.Sprintf("[Player%d]", p))

	i := 0
	for _, r := range s.Game.Field[p] {
		rightText := cardImg(r)

		if s.isSelected(i, p) {
			colorAll(Red, rightText)
		} else if s.cursor.IsSelected(i, p) {
			colorAll(Yellow, rightText) 
		}
		text = concat(text, rightText)
		i++
	}
	for _, r := range s.Perms[p] {
		perm, ok := s.Game.GetPerm(r)
		if !ok {
			continue
		}
		rightText := permImg(perm)
		if s.cursor.IsSelected(i, p) {
			yellow(rightText) 
		}
		text = concat(text, rightText) 
		i++
	}
	
	RightBorder(text)
	return text
}

func (s Screen) gameHeader() []string {
	text := make([]string, 4) 
	text[1] = fmt.Sprintf("%3s", GameTitle)
	text[2] = fmt.Sprintf("Current Player: %d | Mana: %d", 
		s.Game.CurrentPlayer,
		s.Game.Mana)  
	text[3] = fmt.Sprintf("%s", s.Game.AwaitStatus())
	
	return text
}
