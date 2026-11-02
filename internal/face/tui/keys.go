package tui

import (
	"strconv"
	"time"
	"unicode"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
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
	escapeKey        = "esc"
	enterKey         = "enter"
	backspaceKey     = "backspace"
)

// How long a number typed alone waits for a key to count before it chooses the tab of that number
const countPause = 500 * time.Millisecond

// What a digit is worth in the number typed before it
const decimalBase = 10

// mode is what the keys do: move through the page, type a search, type a command, or close the window of keys.
type mode int

const (
	// modeNormal moves through the page and between tabs
	modeNormal mode = iota
	// modeSearch types a search after the slash
	modeSearch
	// modeCommand types a command after the colon
	modeCommand
	// modeKeys shows the window of keys, which the next key closes
	modeKeys
)

// keyState is what the keys pressed so far leave pending: the mode, what is typed on the command line after the
// slash or the colon, a first g, and a number typed before a key, with the sequence of its last digit, so a pause
// that ends an older number is passed over.
type keyState struct {
	mode          mode
	typing        string
	topPending    bool
	count         int
	countSequence int
}

// lineEdit is what a key does to the line the reader types: nothing, a change to it, or leaving it.
type lineEdit int

const (
	lineKept lineEdit = iota
	lineChanged
	lineLeft
)

// countPaused is the pause after a digit, by the digit's sequence.
type countPaused struct {
	sequence int
}

// action is what a key does in normal mode, by the number typed before it, one where none was.
type action func(model Model, count int) (Model, tea.Cmd)

// normalActions are what each key does in normal mode.
func normalActions() map[string]action {
	actions := map[string]action{
		quitKey:          Model.closed,
		interruptKey:     Model.quit,
		downKey:          Model.scrollDown,
		downArrow:        Model.scrollDown,
		upKey:            Model.scrollUp,
		upArrow:          Model.scrollUp,
		topKey:           Model.toLine,
		homeKey:          Model.home,
		bottomKey:        Model.toLineOrEnd,
		endKey:           Model.end,
		nextKey:          Model.nextTab,
		rightArrow:       Model.nextTab,
		tabKey:           Model.nextTab,
		previousKey:      Model.previousTab,
		leftArrow:        Model.previousTab,
		backwardTabKey:   Model.previousTab,
		searchKey:        Model.startSearch,
		commandKey:       Model.startCommand,
		nextMatchKey:     Model.nextMatch,
		previousMatchKey: Model.previousMatch,
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
// adds itself, a backspace takes the last letter back, and Esc, Ctrl+C, as in Neovim, or a backspace on nothing
// leaves the line. Anything else, a control character among it, is passed over, so it never reaches the screen.
func edited(typing string, key tea.KeyPressMsg) (string, lineEdit) {
	switch key.String() {
	case escapeKey, interruptKey:
		return typing, lineLeft
	case backspaceKey:
		if typing == "" {
			return typing, lineLeft
		}
		_, last := utf8.DecodeLastRuneInString(typing)
		return typing[:len(typing)-last], lineChanged
	}
	if !printable(key.Text) {
		return typing, lineKept
	}
	return typing + key.Text, lineChanged
}

// normalKey is the model after a key in normal mode: a digit of a number, a first g, or a key's action, counted
// by the number before it.
func (model Model) normalKey(key string) (Model, tea.Cmd) {
	model.said = commandMessage{text: "", failed: false}
	if digit, isDigit := digitOf(key); isDigit {
		return model.counting(digit)
	}
	topPending := model.keys.topPending
	model.keys.topPending = false
	if key == topKey && !topPending {
		// the number before it waits for the second g, so the pause after its last digit chooses no tab
		model.keys.topPending = true
		model.keys.countSequence++
		return model, nil
	}
	count := model.keys.count
	model.keys.count = 0
	act, known := model.actions[key]
	if !known {
		return model, nil
	}
	return act(model, count)
}

// counting adds a digit to the number typed, and waits a pause for a key to count.
func (model Model) counting(digit int) (Model, tea.Cmd) {
	model.keys.count = model.keys.count*decimalBase + digit
	model.keys.countSequence++
	sequence := model.keys.countSequence
	return model, tea.Tick(countPause, func(time.Time) tea.Msg { return countPaused{sequence: sequence} })
}

// countEnded chooses the tab of a number typed alone, once the pause after its last digit is over.
func (model Model) countEnded(paused countPaused) Model {
	if paused.sequence != model.keys.countSequence || model.keys.count == 0 {
		return model
	}
	place := model.keys.count - firstTabNumber
	model.keys.count = 0
	if place >= len(model.tabs) {
		return model
	}
	return model.choose(place)
}

// digitOf is the digit a key names, where it names one.
func digitOf(key string) (int, bool) {
	digit, err := strconv.Atoi(key)
	return digit, err == nil
}

func (model Model) quit(int) (Model, tea.Cmd) {
	return model, tea.Quit
}

func (model Model) scrollDown(count int) (Model, tea.Cmd) {
	return model.scrolledTo(model.scrolls[model.chosen] + max(count, 1)), nil
}

func (model Model) scrollUp(count int) (Model, tea.Cmd) {
	return model.scrolledTo(model.scrolls[model.chosen] - max(count, 1)), nil
}

// toLine goes to the line of a number, the first where none was typed: gg, and 5gg.
func (model Model) toLine(count int) (Model, tea.Cmd) {
	return model.scrolledTo(max(count, 1) - 1), nil
}

// toLineOrEnd goes to the line of a number, the end where none was typed: G, and 5G.
func (model Model) toLineOrEnd(count int) (Model, tea.Cmd) {
	if count == 0 {
		return model.end(count)
	}
	return model.toLine(count)
}

func (model Model) home(int) (Model, tea.Cmd) {
	return model.scrolledTo(0), nil
}

func (model Model) end(int) (Model, tea.Cmd) {
	return model.scrolledTo(len(model.drawn.tabs[model.chosen].lines)), nil
}

func (model Model) nextTab(count int) (Model, tea.Cmd) {
	return model.choose(model.chosen + max(count, 1)), nil
}

func (model Model) previousTab(count int) (Model, tea.Cmd) {
	return model.choose(model.chosen - max(count, 1)), nil
}

// printable is whether typed text is letters a terminal shows as they are, with no control character among them.
func printable(text string) bool {
	for _, letter := range text {
		if !unicode.IsPrint(letter) {
			return false
		}
	}
	return text != ""
}
