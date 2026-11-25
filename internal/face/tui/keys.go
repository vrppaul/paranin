package tui

import (
	"strconv"
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// The keys a reader presses, as Bubble Tea names them
const (
	quitKey          = "q"
	interruptKey     = "ctrl+c"
	downKey          = "j"
	downArrow        = "down"
	upKey            = "k"
	upArrow          = "up"
	topKey           = "g"
	homeKey          = "home"
	bottomKey        = "G"
	endKey           = "end"
	nextKey          = "l"
	rightArrow       = "right"
	tabKey           = "tab"
	previousKey      = "h"
	leftArrow        = "left"
	backwardTabKey   = "shift+tab"
	searchKey        = "/"
	commandKey       = ":"
	keyWindowKey     = "?"
	spaceKey         = "space"
	nextMatchKey     = "n"
	previousMatchKey = "N"
	copyKey          = "y"
	escapeKey        = "esc"
	enterKey         = "enter"
	backspaceKey     = "backspace"
)

// The most code points the line typed after a slash or a colon holds: far past any place, command or search here,
// and few enough that the work each key does on the line stays small however long a session types.
const typedLimit = 256

// mode is what the keys do: move through the page, type a search, type a command, or close the window of keys.
type mode int

const (
	// modeNormal moves through the page and between tabs
	modeNormal mode = iota
	// modeSearch types a search after the slash
	modeSearch
	// modeCommand types a place to go to, or a command, after the colon
	modeCommand
	// modeKeys shows the window of keys, which the next key closes
	modeKeys
)

// keyState is what the keys pressed so far leave pending: the mode, what is typed on the command line after the
// slash or the colon, and a first g.
type keyState struct {
	mode       mode
	typing     string
	topPending bool
}

// lineEdit is what a key does to the line the reader types: nothing, a change to it, or leaving it.
type lineEdit int

const (
	lineKept lineEdit = iota
	lineChanged
	lineLeft
)

// action is what a key does in normal mode. No number before a key counts it, as the page's vim keys take none.
type action func(model Model) (Model, tea.Cmd)

// normalActions are what each key does in normal mode; gg, the second g's, goes to the top as Home does.
func normalActions() map[string]action {
	actions := map[string]action{
		quitKey:      Model.closed,
		interruptKey: Model.quit,
		downKey:      Model.downward,
		downArrow:    Model.downward,
		upKey:        Model.upward,
		upArrow:      Model.upward,
		enterKey:     Model.followLink,
		topKey:       Model.home,
		homeKey:      Model.home,
		bottomKey:    Model.end,
		endKey:       Model.end,
		nextKey:      Model.nextTab,
		rightArrow:   Model.nextTab,
		tabKey: func(model Model) (Model, tea.Cmd) {
			return model.commandLineOpened().completed(forwardStep), nil
		},
		previousKey: Model.previousTab,
		leftArrow:   Model.previousTab,
		backwardTabKey: func(model Model) (Model, tea.Cmd) {
			return model.commandLineOpened().completed(backwardStep), nil
		},
		searchKey:        Model.startSearch,
		commandKey:       func(model Model) (Model, tea.Cmd) { return model.commandLineOpened(), nil },
		nextMatchKey:     Model.nextMatch,
		previousMatchKey: Model.previousMatch,
		copyKey:          Model.copiedValue,
		escapeKey:        Model.endSearch,
	}
	for _, key := range windowKeys() {
		actions[key] = Model.openKeys
	}
	return actions
}

// pressed is the model after a key, as the mode it is in takes it. Esc typed just before a key reaches a terminal
// program glued to it, as the key with Alt, so such a key is read as Esc and then the key.
func (model Model) pressed(key tea.KeyPressMsg) (Model, tea.Cmd) {
	if key.Mod&tea.ModAlt != 0 {
		escaped, _ := model.pressed(tea.KeyPressMsg{Code: tea.KeyEscape})
		key.Mod &^= tea.ModAlt
		return escaped.pressed(key)
	}
	switch model.keys.mode {
	case modeSearch:
		return model.typedInSearch(key), nil
	case modeCommand:
		return model.typedInCommand(key)
	case modeKeys:
		return model.keyInWindow(key.String())
	case modeNormal:
	}
	return model.normalKey(key.String())
}

// edited is the line the reader types after a key, and what the key did to it: a letter a terminal shows as it is
// adds itself, up to the line's limit, a backspace takes the last letter back, whole as the terminal shows it, and
// Esc, Ctrl+C, as in Neovim, or a backspace on nothing leaves the line. Anything else, a control character among it,
// is passed over, so it never reaches the screen.
func edited(typing string, key tea.KeyPressMsg) (string, lineEdit) {
	switch key.String() {
	case escapeKey, interruptKey:
		return typing, lineLeft
	case backspaceKey:
		if typing == "" {
			return typing, lineLeft
		}
		return withoutLastLetter(typing), lineChanged
	}
	if !printable(key.Text) || utf8.RuneCountInString(typing)+utf8.RuneCountInString(key.Text) > typedLimit {
		return typing, lineKept
	}
	return typing + key.Text, lineChanged
}

// withoutLastLetter is a text with its last letter, a grapheme cluster of one code point or more, taken off.
func withoutLastLetter(text string) string {
	lastStart := 0
	for rest := text; rest != ""; {
		letter, _ := ansi.FirstGraphemeCluster(rest, ansi.GraphemeWidth)
		lastStart, rest = len(text)-len(rest), rest[len(letter):]
	}
	return text[:lastStart]
}

// normalKey is the model after a key in normal mode: a digit choosing its tab, a first g, or a key's action.
func (model Model) normalKey(key string) (Model, tea.Cmd) {
	topPending := model.keys.topPending
	model = model.pendingAndSaidCleared()
	if digit, isDigit := digitOf(key); isDigit {
		return model.numbered(digit), nil
	}
	if key == topKey && !topPending {
		model.keys.topPending = true
		return model, nil
	}
	act, known := model.actions[key]
	if !known {
		return model, nil
	}
	return act(model)
}

// pendingAndSaidCleared is the model as any input in normal mode leaves it, a key, a click or the wheel: a first g
// dropped, and what the command line said cleared.
func (model Model) pendingAndSaidCleared() Model {
	model.said = commandMessage{text: "", failed: false}
	model.keys.topPending = false
	return model
}

// numbered chooses the tab of a digit's number at once, as tmux's prefix and a digit chooses a window; a digit no tab
// is numbered by, zero among them, chooses none.
func (model Model) numbered(digit int) Model {
	place := digit - firstTabNumber
	if place < 0 || place >= len(model.tabs) {
		return model
	}
	return model.choose(place)
}

// digitOf is the digit a key names, where it names one.
func digitOf(key string) (int, bool) {
	digit, err := strconv.Atoi(key)
	return digit, err == nil
}

func (model Model) quit() (Model, tea.Cmd) {
	return model, tea.Quit
}

func (model Model) home() (Model, tea.Cmd) {
	return model.scrolledTo(0), nil
}

func (model Model) end() (Model, tea.Cmd) {
	return model.scrolledTo(len(model.drawn.tabs[model.chosen].lines)), nil
}

func (model Model) nextTab() (Model, tea.Cmd) {
	return model.choose(model.chosen + 1), nil
}

func (model Model) previousTab() (Model, tea.Cmd) {
	return model.choose(model.chosen - 1), nil
}

// shownLetter is a letter of a text from outside as strings.Map keeps it: a letter a terminal shows as it is, any
// other, a control character among them, left out, so the text reaches the screen only as letters, as what the reader
// types does.
func shownLetter(letter rune) rune {
	if unicode.IsPrint(letter) {
		return letter
	}
	return leftOut
}

// What strings.Map leaves a letter out for
const leftOut = -1

// printable is whether typed text is letters a terminal shows as they are, with no control character among them.
func printable(text string) bool {
	for _, letter := range text {
		if !unicode.IsPrint(letter) {
			return false
		}
	}
	return text != ""
}
