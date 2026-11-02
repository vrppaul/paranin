package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// How the command line opens, and what it says, as Neovim says it, of a command it does not know, of a command
// that needs an argument and was given none, and of one that takes none and was given one
const (
	commandMark        = ":"
	unknownCommand     = "E492: Not an editor command: "
	argumentRequired   = "E471: Argument required"
	trailingCharacters = "E488: Trailing characters: "
)

// What the window of commands says the quit command does on the résumé, which the status line and the window of
// keys say of q there
const quitHint = "quit"

// A command runs with no count typed before it
const noCount = 0

// command is a command the command line runs: its name; what the window of commands says it does, or nothing
// where the window does not list it, as for a second name of a command it lists; whether it takes an argument,
// which it needs then; and what it does with the argument.
type command struct {
	name          string
	does          string
	takesArgument bool
	act           commandAction
}

// commandAction is what a command does, with the argument typed after it, or nothing where it takes none.
type commandAction func(model Model, argument string) (Model, tea.Cmd)

// commands are the commands the command line runs, in the order the window lists them, which says what q does on
// the page shown: Neovim's ways to close a window and to quit, listed by the first alone, and its way to open a
// file, here a project's page by its slug.
func commands(closing string) []command {
	return []command{
		{name: "q", does: closing, takesArgument: false, act: Model.closedByCommand},
		{name: "quit", does: "", takesArgument: false, act: Model.closedByCommand},
		{name: "q!", does: "", takesArgument: false, act: Model.closedByCommand},
		{name: "qa", does: "", takesArgument: false, act: Model.quitByCommand},
		{name: "e", does: "open a page", takesArgument: true, act: Model.edit},
		{name: "edit", does: "", takesArgument: true, act: Model.edit},
	}
}

// closedByCommand closes the page shown, as q does.
func (model Model) closedByCommand(string) (Model, tea.Cmd) {
	return model.closed(noCount)
}

// quitByCommand ends the session, as Ctrl+C does.
func (model Model) quitByCommand(string) (Model, tea.Cmd) {
	return model.quit(noCount)
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

// ranCommand runs the command typed, as Neovim does: the one named by the first word, with what follows it, or none,
// and the command line says it knows no such command; the keys go back to normal mode either way, and nothing typed
// runs nothing.
func (model Model) ranCommand() (Model, tea.Cmd) {
	model.keys.mode = modeNormal
	typed := strings.TrimSpace(model.keys.typing)
	if typed == "" {
		return model, nil
	}
	name, argument, _ := strings.Cut(typed, space)
	for _, known := range model.commands {
		if known.name == name {
			return model.ranWith(known, strings.TrimSpace(argument))
		}
	}
	model.said = commandMessage{text: unknownCommand + typed, failed: true}
	return model, nil
}

// ranWith runs a command with the argument typed after it, where the command takes that: one that takes an argument
// needs it, and one that takes none refuses any, each saying so as Neovim says it.
func (model Model) ranWith(known command, argument string) (Model, tea.Cmd) {
	switch {
	case known.takesArgument && argument == "":
		model.said = commandMessage{text: argumentRequired, failed: true}
	case !known.takesArgument && argument != "":
		model.said = commandMessage{text: trailingCharacters + argument, failed: true}
	default:
		return known.act(model, argument)
	}
	return model, nil
}
