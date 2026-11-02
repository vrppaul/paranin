package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// How the command line opens, and what it says of a command it does not know, as Neovim says it
const (
	commandMark    = ":"
	unknownCommand = "E492: Not an editor command: "
)

// What the window of commands says the quit command does, which the status line and the window of keys say of q
const quitHint = "quit"

// A command runs with no count typed before it
const noCount = 0

// command is a command the command line runs: its name; what the window of commands says it does, or nothing
// where the window does not list it, as for a second name of a command it lists; and what it does.
type command struct {
	name string
	does string
	act  action
}

// commands are the commands the command line runs, in the order the window lists them: Neovim's ways to quit
// among them, listed by the first alone.
func commands() []command {
	return []command{
		{name: "q", does: quitHint, act: Model.quit},
		{name: "quit", does: "", act: Model.quit},
		{name: "qa", does: "", act: Model.quit},
		{name: "q!", does: "", act: Model.quit},
	}
}

// completionState is the window of commands: the commands it lists, worked out each time the reader edits the
// line, and whether Tab has chosen the first of them.
type completionState struct {
	listed []command
	chosen bool
}

// startCommand opens the command line, empty, after a colon; the window lists every command.
func (model Model) startCommand(int) (Model, tea.Cmd) {
	model.keys.mode, model.keys.typing = modeCommand, ""
	model.completion = completionState{listed: model.completions(""), chosen: false}
	return model, nil
}

// typedInCommand is the model after a key while the reader types a command: Enter runs it, Tab chooses the first
// command the window lists; any other key edits the line, and the window lists anew, or leaves it unrun.
func (model Model) typedInCommand(key tea.KeyPressMsg) (Model, tea.Cmd) {
	switch key.String() {
	case enterKey:
		return model.ranCommand()
	case tabKey:
		return model.completed(), nil
	}
	typing, edit := edited(model.keys.typing, key)
	switch edit {
	case lineChanged:
		model.keys.typing = typing
		model.completion = completionState{listed: model.completions(typing), chosen: false}
	case lineLeft:
		model.keys.mode = modeNormal
	case lineKept:
	}
	return model, nil
}

// completed is the model with the first command the window lists chosen and written on the command line.
func (model Model) completed() Model {
	if len(model.completion.listed) == 0 {
		return model
	}
	model.keys.typing, model.completion.chosen = model.completion.listed[0].name, true
	return model
}

// completions are the commands the window lists for what the reader typed: those it lists at all whose names start
// with it.
func (model Model) completions(typed string) []command {
	listed := []command{}
	for _, known := range model.commands {
		if known.does != "" && strings.HasPrefix(known.name, typed) {
			listed = append(listed, known)
		}
	}
	return listed
}

// ranCommand runs the command typed, as Neovim does: the one of that name, or none, and the command line says it
// knows no such command; the keys go back to normal mode either way, and nothing typed runs nothing.
func (model Model) ranCommand() (Model, tea.Cmd) {
	model.keys.mode = modeNormal
	typed := strings.TrimSpace(model.keys.typing)
	if typed == "" {
		return model, nil
	}
	for _, known := range model.commands {
		if known.name == typed {
			return known.act(model, noCount)
		}
	}
	model.said = commandMessage{text: unknownCommand + typed, failed: true}
	return model, nil
}
