package tui

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/vrppaul/paranin/internal/document"
)

// What the prompt says, as Neovim's command line says it, of a command it does not know, of a command that needs an
// argument and was given none, and of one that takes none and was given one
const (
	unknownCommand     = "E492: Not an editor command: "
	argumentRequired   = "E471: Argument required"
	trailingCharacters = "E488: Trailing characters: "
)

// What the window of keys says q does on the résumé
const quitHint = "quit"

// The command that opens a project's page, as Neovim's opens a file
const editCommand = "e"

// hiddenCommand is a command that shows a page the session may hold, kept from the places the prompt lists, as the
// first version kept it hidden: its name, and the page.
type hiddenCommand struct {
	name string
	page document.Page
}

// hiddenCommands are the commands that show the hidden pages: Neovim's :checkhealth its health report, and
// lazy.nvim's :Lazy its list of plugins, here the list of skills.
func hiddenCommands() []hiddenCommand {
	return []hiddenCommand{
		{name: "checkhealth", page: document.Page{Kind: document.PageHealth, Slug: ""}},
		{name: "Lazy", page: document.Page{Kind: document.PageSkills, Slug: ""}},
	}
}

// A command runs with no count typed before it
const noCount = 0

// command is a command the prompt runs, as Neovim's command line does, though it lists none of them: the names it
// goes by; whether it takes an argument, which it needs then; and what it does with the argument.
type command struct {
	names         []string
	takesArgument bool
	act           commandAction
}

// commandAction is what a command does, with the argument typed after it, or nothing where it takes none.
type commandAction func(model Model, argument string) (Model, tea.Cmd)

// commands are the commands the prompt runs: Neovim's ways to close a window, as q closes it, and to quit, as Ctrl+C
// does, and its way to open a file, here a project's page by its slug; then the hidden commands of the hidden pages
// the session holds, each showing its page in place of the one shown, as :e does a project's.
func commands(buffers map[document.Page][]tab) []command {
	known := []command{
		{names: []string{"q", "quit", "q!"}, takesArgument: false, act: func(model Model, _ string) (Model, tea.Cmd) { return model.closed(noCount) }},
		{names: []string{"qa"}, takesArgument: false, act: func(model Model, _ string) (Model, tea.Cmd) { return model.quit(noCount) }},
		{names: []string{editCommand, "edit"}, takesArgument: true, act: Model.edit},
	}
	for _, hidden := range hiddenCommands() {
		if _, held := buffers[hidden.page]; held {
			known = append(known, command{names: []string{hidden.name}, takesArgument: false, act: func(model Model, _ string) (Model, tea.Cmd) { return model.showing(hidden.page), nil }})
		}
	}
	return known
}

// completion is an item the places under the prompt list: what it shows, and what Tab writes on the prompt where it
// chooses the item.
type completion struct {
	shown   string
	written string
}

// completionState is what the places under the prompt list: what the reader typed, what is listed for it, both
// worked out each time they edit the line, and the place of the item Tab or Shift+Tab has chosen.
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

// promptOpened is the model with the prompt open, empty, as a colon opens it; the places listed are those listed
// before anything is typed.
func (model Model) promptOpened() Model {
	model.keys.mode, model.keys.typing = modeCommand, ""
	model.completion = model.completionsOf("")
	return model
}

// completionsOf is what the places list for what the reader typed, none of it chosen.
func (model Model) completionsOf(typed string) completionState {
	return completionState{typed: typed, listed: model.completions(typed), chosen: noneChosen}
}

// typedInCommand is the model after a key while the reader types a place or a command: Enter goes there or runs it,
// Tab and Shift+Tab go round the places listed; any other key edits the line, and the places are listed anew, or
// leaves it unrun.
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

// completed is the model with the item a step on chosen and written on the prompt, as Neovim's wildmenu goes
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

// commandParts are what the prompt holds, read as Neovim reads its command line: a command's name, then, after a
// space, its argument, spaces before either left out; and whether a space came after the name.
func commandParts(typed string) (name, argument string, argued bool) {
	name, argument, argued = strings.Cut(strings.TrimLeft(typed, space), space)
	return name, strings.TrimLeft(argument, space), argued
}

// completions are what the prompt lists for what the reader typed: before a space, the places whose names start with
// it; after a command and a space, the pages whose slugs start with what follows, where the command takes an
// argument, which only a page is.
func (model Model) completions(typed string) []completion {
	name, argument, argued := commandParts(typed)
	if argued {
		return model.pageCompletions(name, argument)
	}
	return model.placesFor(model.chosen, name)
}

// pageCompletions are the projects' pages the prompt lists after a command named and a space, by slug, where the
// command takes an argument: those whose slugs start with what is typed, each written after the command's name.
func (model Model) pageCompletions(name, typed string) []completion {
	listed := []completion{}
	if known, found := model.commandNamed(name); !found || !known.takesArgument {
		return listed
	}
	for _, slug := range model.slugs {
		if strings.HasPrefix(slug, typed) {
			listed = append(listed, completion{shown: slug, written: name + space + slug})
		}
	}
	return listed
}

// commandNamed is the command of a name, and whether there is one.
func (model Model) commandNamed(name string) (command, bool) {
	for _, known := range model.commands {
		if slices.Contains(known.names, name) {
			return known, true
		}
	}
	return command{names: nil, takesArgument: false, act: nil}, false
}

// ranCommand goes to the place typed, or runs the command typed, as Neovim does: the one named by the first word,
// with what follows it, or none, and the prompt says it knows no such command; the keys go back to normal mode
// either way, and nothing typed runs nothing.
func (model Model) ranCommand() (Model, tea.Cmd) {
	model.keys.mode = modeNormal
	typed := strings.TrimSpace(model.keys.typing)
	if typed == "" {
		return model, nil
	}
	if known, found := model.placeNamed(typed); found {
		return known.act(model, noCount)
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
