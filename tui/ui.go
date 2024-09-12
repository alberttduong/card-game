package tui 

import (
	"github.com/alberttduong/card-game/game"
	"strings"
	"slices"
	"fmt"
	"strconv"
	"unicode/utf8"
	"errors"
)

const (
	ColorOffset = 11
	CardHeight = 5
	CardWidth = 5
	FieldWithoutPermsWidth = FieldLength - (CardWidth * 3 + 2) - FieldHeaderWidth
	FieldHeight = CardHeight 
	FieldHeaderWidth = 10
	FieldLength = 50 
	RenderPadding = 4

	GameTitle = "Wizard Spell Battle Fight the Card Game"
)

type TextWrapIter struct {
	lines []string
	current int
	width int
}

func NewTextWrapIter(w int) *TextWrapIter {
	return &TextWrapIter{width: w}
}

func (t *TextWrapIter) AddLines(lines ...string) error {
	for _, line := range lines {
		if runeLen(line) > t.width {
			return errors.New("Line too long")
		}
		t.lines = append(t.lines, line)
	}
	return nil
}

func (t *TextWrapIter) AddParagraph(p string) {
	t.lines = slices.Concat(t.lines, wrapText(p, t.width))
}

func (t *TextWrapIter) AddWizardAttackDesc(card game.Cdata) {
	addAtk := func (a game.Attack) {	
		empty := game.Attack{}
		if a == empty { 
			return
		}
		name := "[" + a.Name + "]" 
		damage := fmt.Sprintf("%d󰓥", a.Dmg)
		title := middlePad(t.width, name, damage) 
		t.AddLines(title)
		t.AddParagraph(a.Desc)
		t.AddLines("")
	}
	addAtk(card.Atk0)
	addAtk(card.Atk1)
}

func (t *TextWrapIter) Next() (string, error) {
	if t.current >= len(t.lines) {
		return "", errors.New("End") 
	}
	defer func() {
		t.current++
	}()
	return t.lines[t.current], nil
}

func (t *TextWrapIter) ToStrings() []string {
	return t.lines
}

func wrapText(s string, width int) (lines []string) {
	newLine := ""
	for _, word := range strings.Split(s, " ") {
		if runeLen(newLine) + runeLen(word) > width {
			lines = append(lines, strings.TrimSuffix(newLine, " "))
			newLine = ""
		}
		newLine += word + " "
	}
	
	return append(lines, strings.TrimSuffix(newLine, " "))
}

func cardNameImg(cards []game.Cdata, c game.CardName) []string {
	data := cards[c - 1]
	name := fmt.Sprint(data.CName)
	return []string {
		"┌───┐",
		fmt.Sprintf("│%s│", name[0:3]),
		fmt.Sprintf("│%s│", fmt.Sprintf("%-3s", name[3:])[:3]),
		fmt.Sprintf("│%2d󰓏│", data.Hp), 
		"└───┘",
	}
}

func permImg(c game.Perm) []string {
	return []string{
		"┌───┐",
		fmt.Sprintf("│%s│", fmt.Sprint(c.CName)[0:3]),
		boxMiddle(3, ""),
		boxMiddle(3, ""),
		"└───┘",
	}
}

func cardImg(c game.Card) []string {
	dmg0 := strings.TrimPrefix(strconv.Itoa(c.Atk0.Dmg), "-")
	return []string{
		"┌───┐",
		fmt.Sprintf("│%s│", fmt.Sprint(c.CName)[0:3]),
		fmt.Sprintf("│%2d│", c.HP),
		fmt.Sprintf("│%s󰓥%d│", dmg0, c.Atk1.Dmg),
		"└───┘",
	}
}

func render(s []string) {
	for _, line := range s {
		fmt.Println(pad2(line))
	}
}

func pad2(s string) string {
	return fmt.Sprint(strings.Repeat(" ", RenderPadding), s)
}

func fieldHeader(lines ...string) []string {
	text := make([]string, FieldHeight) 
	content := NewTextWrapIter(FieldHeaderWidth)
	content.lines = lines
	for i := range FieldHeight {
		line, _ := content.Next()
		text[i] = fmt.Sprintf("│%-*s│", FieldHeaderWidth-1, line) 
	}
	return text 
}

func RightBorder(text []string) {
	for i, s := range text {
		text[i] = fmt.Sprintf("%s%s│",
			s,
			strings.Repeat(" ", FieldLength - runeLen(s) + 1),
		)
	}
}

func fieldLineWithTopCross() []string {
	return fieldLineWithSep("┬", "┌", "┐") 
}

func fieldLineWithBottomCross() []string {
	return fieldLineWithSep("┴", "└", "┘") 
}

func fieldLineWithCross() []string {
	return fieldLineWithSep("┼", "├", "┤") 
}

func fieldLineWithSep(sep, first, last string) []string {
	s := []string {
		fmt.Sprint(
			first,
			strings.Repeat("─", FieldHeaderWidth - 1),
			sep, 
			strings.Repeat("─", FieldLength - FieldHeaderWidth),
			last,
		),
	}
	return s
}

func horzFieldLine() []string {
	return []string {
		strings.Repeat("─", FieldLength),	
	}
}

func green(s []string) []string {
	for i, line := range s {
		s[i] = fmt.Sprintf("\x1b[1;32m%s\x1b[0m", line)// + "\x1b[0m"
	}
	return s
}

type colorCode int
const (
	Green colorCode = 32
	Yellow colorCode = 33
	Red colorCode = 31
)

func colorAll(c colorCode, s []string) []string {
	for i, l := range s {
		s[i] = fmt.Sprintf("\x1b[1;%dm%s\x1b[0m", c, l)
	}
	return s
}

func color(c colorCode, s string) string {
	return fmt.Sprintf("\x1b[1;%dm%s\x1b[0m", c, s)
}

func yellow(s []string) []string {
	for i, line := range s {
		s[i] = color(Yellow, line)
	}
	return s
}

func boxMiddle(width int, s string) string {
	return fmt.Sprintf("│%*s│", width, s) 
}

func boxMiddleLine(width int) string {
	return fmt.Sprintf("├%s┤", strings.Repeat("─", width)) 
}

func boxLeftJustifyMiddle(width int, s string) string {
	return fmt.Sprintf("│%-*s│", width, s) 
}

func boxTop(width int) string {
	return fmt.Sprintf("┌%s┐", strings.Repeat("─", width))
}

func boxBottom(width int) string {
	return fmt.Sprintf("└%s┘", strings.Repeat("─", width))
}

func runeLen(s string) int {
	numColors := 0
	copyS := strings.Clone(s)
	i := strings.Index(copyS, "[0m")
	for i != -1 {
		copyS = copyS[i+1:]
		numColors++
		i = strings.Index(copyS, "[0m")
	}
	
	return utf8.RuneCountInString(s) - numColors * ColorOffset
}

func maxWidth(s []string) (maxW int) {
	for _, line := range s {
		if l := runeLen(line); l > maxW {
			maxW = l
		}
	}
	return maxW
}

func leftPad(width int, s string) string {
	pad := width - runeLen(s)
	if pad < 0 {
		pad = 0
	}

	return fmt.Sprint(strings.Repeat(" ", pad), s)
}

func clearScreen() {
	fmt.Print("\033[H\r")
	fmt.Print("\033[2J\r")	
}

func middlePad(width int, l, r string) string {
	pad := width - runeLen(l) - runeLen(r)
	if pad < 0 {
		pad = 0
	}

	return fmt.Sprint(l, strings.Repeat(" ", pad), r)
}
func rightPad(width int, s string) string {
	pad := width - runeLen(s)
	if pad < 0 {
		pad = 0
	}

	return fmt.Sprint(s, strings.Repeat(" ", pad))
}

func concatMany(texts ...[]string) (r []string) {
	for _, s := range texts {
		r = concat(r, s)
	}
	return r
}

func concat(l, r []string) []string {
	lenLeft, lenRight := len(l), len(r)
	var i, j int

	left := slices.Clone(l)
	right := slices.Clone(r)
	
	leftWidth := maxWidth(left)
	for i < lenLeft && j < lenRight { 
		//
		//left[i] = fmt.Sprintf("%-*s", leftWidth, left[i]) + right[i]
		left[i] = rightPad(leftWidth, left[i]) + right[i]
		i++
		j++
	}
	for ; j < lenRight; j++ {
		left = append(left, strings.Repeat(" ", leftWidth) + right[j])
	}

	return left
}

func Box(text string) []string {
	return []string{
		boxTop(len(text)),
		boxMiddle(len(text), text),
		boxBottom(len(text)),
	}
}
