package compose

import (
	"slices"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"

	"github.com/vrppaul/paranin/internal/document"
)

// The language of code whose author names none: a project's recording, a post's code indented
const noLanguage = ""

// lexedLanguage is a language compose splits code into kinds of word in: chroma's lexer of it, by its name, and the
// names a post's author writes the language by, in small letters.
type lexedLanguage struct {
	lexer string
	names []string
}

// codeBlock is code in runs of its kinds of word, as chroma's lexer of its language reads them; one plain run in a
// language lexedLanguages names none of, or where its lexer fails, so the code is still shown whole.
func codeBlock(text, language string) document.CodeBlock {
	plain := document.CodeBlock{Language: language, Runs: []document.CodeRun{{Kind: document.CodePlain, Text: text}}}
	named := strings.ToLower(language)
	languages := lexedLanguages()
	found := slices.IndexFunc(languages, func(lexed lexedLanguage) bool { return slices.Contains(lexed.names, named) })
	if found < 0 {
		return plain
	}
	lexer := lexers.Get(languages[found].lexer)
	if lexer == nil {
		return plain
	}
	tokens, err := lexer.Tokenise(nil, text)
	if err != nil {
		return plain
	}
	kinds := codeKinds()
	runs := []document.CodeRun{}
	unread := len(text)
	for _, token := range tokens.Tokens() {
		// a lexer may end the text with a line's end the block does not hold (EnsureNL)
		written := token.Value[:min(len(token.Value), unread)]
		unread -= len(written)
		runs = appendRun(runs, document.CodeRun{Kind: codeKindOf(kinds, token.Type), Text: written})
	}
	return document.CodeBlock{Language: language, Runs: runs}
}

// lexedLanguages are the languages a post's code is split into kinds of word in. Any other stays plain: a page is
// drawn on the first request after a snapshot, so a lexer that never ends would hang its readers, and only lexers a
// test runs on code of their language are trusted. A diff is left out, since its lines are of no kind a face colours.
func lexedLanguages() []lexedLanguage {
	return []lexedLanguage{
		{lexer: "Go", names: []string{"go"}},
		{lexer: "Bash", names: []string{"sh", "bash", "shell"}},
		{lexer: "Bash Session", names: []string{"console", "shell-session"}},
		{lexer: "YAML", names: []string{"yaml", "yml"}},
		{lexer: "JSON", names: []string{"json"}},
		{lexer: "TOML", names: []string{"toml"}},
		{lexer: "Python", names: []string{"python", "py"}},
		{lexer: "Rust", names: []string{"rust", "rs"}},
		{lexer: "TypeScript", names: []string{"typescript", "ts"}},
		{lexer: "JavaScript", names: []string{"javascript", "js"}},
		{lexer: "SQL", names: []string{"sql"}},
		{lexer: "Docker", names: []string{"dockerfile"}},
		{lexer: "Makefile", names: []string{"makefile", "make"}},
		{lexer: "HTML", names: []string{"html"}},
		{lexer: "CSS", names: []string{"css"}},
		{lexer: "C", names: []string{"c"}},
		// a widget's own language, which a post about one shows
		{lexer: "Svelte", names: []string{"svelte"}},
	}
}

// appendRun adds a run to runs, joined to the last where the two are of one kind; a run of no text adds nothing.
func appendRun(runs []document.CodeRun, run document.CodeRun) []document.CodeRun {
	if run.Text == "" {
		return runs
	}
	if last := len(runs) - 1; last >= 0 && runs[last].Kind == run.Kind {
		runs[last].Text += run.Text
		return runs
	}
	return append(runs, run)
}

// codeKindOf is the kind of word a type of chroma's token is: its own among the kinds, else its nearest parent's
// there, as chroma's styles inherit a colour, a string's double quotes a string's; plain where none is.
func codeKindOf(kinds map[chroma.TokenType]document.CodeKind, tokenType chroma.TokenType) document.CodeKind {
	for named := tokenType; ; named = named.Parent() {
		if kind, known := kinds[named]; known {
			return kind
		}
		if named == chroma.EOFType {
			return document.CodePlain
		}
	}
}

// codeKinds names the kind of word each type of chroma's token is that a face colours, a parent standing for every
// type under it; every other type, a name a program gives among them, is plain.
func codeKinds() map[chroma.TokenType]document.CodeKind {
	return map[chroma.TokenType]document.CodeKind{ //nolint:exhaustive // chroma's types stand under parents; one left out is its parent's kind
		chroma.Keyword:         document.CodeKeyword,
		chroma.KeywordConstant: document.CodeNumber,
		chroma.KeywordType:     document.CodeType,
		chroma.NameClass:       document.CodeType,
		chroma.NameConstant:    document.CodeNumber,
		chroma.NameBuiltin:     document.CodeBuiltin,
		chroma.NameFunction:    document.CodeFunction,
		// a key of data, as YAML's
		chroma.NameTag:       document.CodeFunction,
		chroma.NameVariable:  document.CodeVariable,
		chroma.LiteralString: document.CodeString,
		// a date written into data, as YAML's, is text written in as a string is
		chroma.LiteralDate:   document.CodeString,
		chroma.LiteralNumber: document.CodeNumber,
		chroma.Operator:      document.CodeOperator,
		// an operator written as a word, as Python's and, reads among the words round it
		chroma.OperatorWord: document.CodePlain,
		chroma.Punctuation:  document.CodePunctuation,
		chroma.Comment:      document.CodeComment,
	}
}
