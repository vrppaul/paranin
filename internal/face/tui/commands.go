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

// The command that opens a project's page, as Neovim's opens a file, and what the status line says it opens
const (
	editCommand = "e"
	pagesHint   = "pages"
)

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
		{name: editCommand, does: "open a page", takesArgument: true, act: Model.edit},
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

// completion is an item the window of commands lists: what it shows, what it says the item does, and what Tab writes
// on the command line where it chooses the item.
type completion struct {
	shown   string
	does    string
	written string
}

// completionState is the window of commands: what the reader typed, what the window lists for it, both worked out
// each time they edit the line, and the place of the item Tab or Shift+Tab has chosen.
type completionState struct {
	typed  string
	listed []completion
	chosen int
}

// The place of the item chosen where none is, the line holding what the reader typed; the places what was typed
// takes in the round Tab goes, its first, the items after it; and the steps Tab and Shift+Tab take through it
const (
	noneChosen   = -1
	typedPlaces  = 1
	forwardStep  = 1
	backwardStep = -1
)

// startCommand opens the command line, empty, after a colon; the window lists every command.
func (model Model) startCommand(int) (Model, tea.Cmd) {
	model.keys.mode, model.keys.typing = modeCommand, ""
	model.completion = model.completionsOf("")
	return model, nil
}

// completionsOf is the window of commands for what the reader typed, none of it chosen.
func (model Model) completionsOf(typed string) completionState {
	return completionState{typed: typed, listed: model.completions(typed), chosen: noneChosen}
}

// typedInCommand is the model after a key while the reader types a command: Enter runs it, Tab and Shift+Tab go
// round what the window lists; any other key edits the line, and the window lists anew, or leaves it unrun.
func (model Model) typedInCommand(key tea.KeyPressMsg) (Model, tea.Cmd) {
	switch key.String() {
	case enterKey:
		return model.ranCommand()
	case tabKey:
		return model.completed(forwardStep), nil
	case backwardTabKey:
		return model.completed(backwardStep), nil
	}
	typing, edit := edited(model.keys.typing, key)
	switch edit {
	case lineChanged:
		model.keys.typing = typing
		model.completion = model.completionsOf(typing)
	case lineLeft:
		model.keys.mode = modeNormal
	case lineKept:
	}
	return model, nil
}

// completed is the model with the item a step on chosen and written on the command line, as Neovim's wildmenu goes
// round: what the reader typed is a place in the round, before the first and after the last, so Tab from it
// chooses the first, Shift+Tab the last, and a step past either end writes what they typed again.
func (model Model) completed(step int) Model {
	listed := model.completion.listed
	if len(listed) == 0 {
		return model
	}
	next := wrapped(model.completion.chosen+typedPlaces+step, len(listed)+typedPlaces) - typedPlaces
	model.keys.typing, model.completion.chosen = model.completion.typed, next
	if next != noneChosen {
		model.keys.typing = listed[next].written
	}
	return model
}

// commandParts are what a command line holds, read as Neovim reads it: a command's name, then, after a space, its
// argument, spaces before either left out; and whether a space came after the name.
func commandParts(typed string) (name, argument string, argued bool) {
	name, argument, argued = strings.Cut(strings.TrimLeft(typed, space), space)
	return name, strings.TrimLeft(argument, space), argued
}

// completions are what the window lists for what the reader typed: before a space, the commands it lists at all
// whose names start with it; after a command and a space, the pages whose slugs start with what follows, where the
// command takes an argument, which only a page is.
func (model Model) completions(typed string) []completion {
	name, argument, argued := commandParts(typed)
	if argued {
		return model.pageCompletions(name, argument)
	}
	listed := []completion{}
	for _, known := range model.commands {
		if known.does != "" && strings.HasPrefix(known.name, name) {
			listed = append(listed, completion{shown: commandMark + known.name, does: known.does, written: known.name})
		}
	}
	return listed
}

// pageCompletions are the projects' pages the window lists after a command named and a space, by slug, where the
// command takes an argument: those whose slugs start with what is typed, each written after the command's name.
func (model Model) pageCompletions(name, typed string) []completion {
	listed := []completion{}
	if known, found := model.commandNamed(name); !found || !known.takesArgument {
		return listed
	}
	for _, slug := range model.slugs {
		if strings.HasPrefix(slug, typed) {
			listed = append(listed, completion{shown: slug, does: "", written: name + space + slug})
		}
	}
	return listed
}

// commandNamed is the command of a name, and whether there is one.
func (model Model) commandNamed(name string) (command, bool) {
	for _, known := range model.commands {
		if known.name == name {
			return known, true
		}
	}
	return command{name: "", does: "", takesArgument: false, act: nil}, false
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
	name, argument, _ := commandParts(typed)
	if known, found := model.commandNamed(name); found {
		return model.ranWith(known, strings.TrimSpace(argument))
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
