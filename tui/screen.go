package tui

import (
	"github.com/alberttduong/card-game/game"
	"github.com/nsf/termbox-go"
	"fmt"
	"slices"
)

type Mode int
const (
	Game Mode = iota
	Deck
	Start

	DefaultMode = Start 
)
var (
	StartDeck = ScreenErr{"Go to deckbuilder screen"}
	StartGame = ScreenErr{"Go to game screen"}
)

type Screener interface {
	Redraw()
	HandleEvent(termbox.Event) error
	Cursor() *Cursor
}

type MainScreen struct {
	lastError error
	CurrentMode Mode
	Current Screener
	Deck *DeckBuilder	
	Game *Screen
	Start *StartScreen
}

func InitScreen(cards []game.Cdata) *MainScreen {
	m := MainScreen{
		CurrentMode: DefaultMode,
		Start: NewStartScreen(),
		Deck: NewDeckBuilder(cards),
		Game: NewScreen(cards), 
	}
	m.SetMode(DefaultMode)
	return &m
}

func (m *MainScreen) SetMode(mode Mode) {
	switch mode {
	case Game:
		g, err := m.Game.InitGame()
		if err != nil {
			m.lastError = err
			break
		}
		m.Current = m.Game
		m.Game.Game = g 
	case Deck:
		m.Current = m.Deck
	default:
		m.Current = m.Start
	}
	m.CurrentMode = mode
	m.Current.Cursor().ResetCursor()
	m.Redraw()
}

func HandleMovement(s Screener, ev termbox.Event) error {
	switch ev.Ch {
	case 'k':
		s.Cursor().Up()
	case 'j':
		s.Cursor().Down()
	case 'h':
		s.Cursor().Left()
	case 'l':
		s.Cursor().Right()
	}

	if ev.Key == termbox.KeyCtrlC {
		return EXIT 
	}
	s.Redraw()
	return nil 
}

func (m *MainScreen) Redraw() {
	m.Current.Redraw()
	if m.lastError == nil {
		return
	}
	render([]string{
		"",
		"Invalid Deck(s):" + m.lastError.Error()})
}

func (m *MainScreen) HandleEvent(ev termbox.Event) error {
	if ev.Key == termbox.KeyCtrlC {
		return EXIT 
	}
	key := ev.Ch
	switch m.CurrentMode {
	case Start:
		if key == '1' {
			m.SetMode(Game)	
			return nil
		} else if key == '2' {
			m.SetMode(Deck)	
			return nil
		}
	}

	HandleMovement(m.Current, ev)

	//TODO
	err := m.Current.HandleEvent(ev)

	if err == BACK {
		m.SetMode(Start)
	} else if err == StartDeck {
		m.SetMode(Deck)
	} else if err == StartGame {
		m.SetMode(Game)
	}
	return nil
}

type Cursor struct {
	Hovered, Selected CardPos
	Coords []Coord
	maxX int
}

func (c Cursor) TargetPermStr(numWizs int) string {
	return fmt.Sprintf("%d %d", c.Selected.y, c.Selected.x - numWizs)
}

func (c Cursor) TargetStr() string {
	return fmt.Sprintf("%d %d", c.Selected.y, c.Selected.x)
	//return fmt.Sprintf("%d %d", c.Selected.y, c.Selected.x)
}

func (s *Cursor) updateX() {
	if s.Selected.y >= len(s.Coords) {
		return
	}
	if minX := s.Coords[s.Selected.y].length - 1; minX < s.maxX {
		s.Selected.x = minX
	}
}

func (s Cursor) IsSelected(x, y int) bool {
	return s.Selected.x == x && 
	       s.Coords[s.Selected.y].realRow == y
}

func (s Cursor) SelectedY() int {
	if s.Selected.y < 0 || s.Selected.y >= len(s.Coords) {
		return -1
	}
	return s.Coords[s.Selected.y].realRow
}

func (s Cursor) SelectedYis(y int) bool {
	if s.Selected.y >= len(s.Coords) {
		return false
	}
	return s.Coords[s.Selected.y].realRow == y
}

func (s *Cursor) Up() {
	if s.Selected.y > 0 {
		s.Selected.y--
		s.updateX()
	}
}

func (s *Cursor) Down() {
	if s.Selected.y < len(s.Coords)-1 {
		s.Selected.y++
		fmt.Println(s.Selected.y)
		s.updateX()
	}
}

func (s *Cursor) Left() {
	if s.Selected.x > 0 {
		s.Selected.x--
		s.maxX = s.Selected.x
	}
}

func (s *Cursor) Right() {
	if s.Selected.y >= len(s.Coords) {
		return
	}
	if s.Selected.x < s.Coords[s.Selected.y].length - 1 { 
		s.Selected.x++
		s.maxX = s.Selected.x
	}
}

func (s *Cursor) ResetCursor() {
	s.Selected = CardPos{}
}

func (s Screen) isSelected(x, y int) bool {
	c := CardPos{x, y}
	return s.selected == c
}

func (s *Screen) resetSelected() {
	s.selected = CardPos{-1, -1}
}

type StartScreen struct {
	cursor *Cursor
}

func (s *StartScreen) HandleEvent(ev termbox.Event) error {
	if ev.Key == termbox.KeyEnter {
		if s.cursor.IsSelected(0, 0) {
			return StartGame
		} else if s.cursor.IsSelected(0, 1) {
			return StartDeck
		}
	}
	return nil
}

func NewStartScreen() *StartScreen {
	return &StartScreen {
		cursor: &Cursor{Coords: []Coord{
			{0, 1},
			{1, 1},
		}}, 
	}
}

func (s *StartScreen) Redraw() {
	clearScreen()
	p := RenderPadding + (len(GameTitle)+len("Start Game")-6)/2
 	start := Box("Start Game")
	for i, s := range start {
		start[i] = leftPad(p, s)
	}
	p = RenderPadding + (len(GameTitle)+len("Deck Builder")-6)/2
	deck := Box("Deck Builder")
	for i, s := range deck {
		deck[i] = leftPad(p, s)
	}
	if s.cursor.IsSelected(0, 0) {
		start = yellow(start)
	} else if s.cursor.IsSelected(0, 1) {
		deck = yellow(deck)
	}

	text := slices.Concat(
		[]string{
			"",
			GameTitle,
			"",
		},
		start,
		deck,
		[]string{
			"Keys:",
			"h, j, k, l",
			"Enter: Select/Play",
			"n: Small attack", "m: Big attack",
			"i: Open Chat, Ctrl-Q: Close Chat",
			"b: Main Menu", 
		},
	)
	render(text)
}

func (s *StartScreen) Cursor() *Cursor {
	return s.cursor
}
