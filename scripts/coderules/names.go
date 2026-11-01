package main

import (
	"go/ast"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"
)

// The fewest letters a word of a name may have and still be a word: `row`, `end`, `key`
const wholeWordLetters = 3

// What parts the words of a name written with underscores
const underscore = "_"

// The small s that makes an initialism plural: `IDs`, `URLs`
const plural = "s"

const (
	cutWordMessage   = "%q is not a whole word (in %q): name what it holds"
	mixedCaseMessage = "%q spells an initialism in mixed case: write %q"
)

const wholeNamesDescription = `refuse a name that is not made of whole words

Every name the code declares is read, in the order the files hold them: the package, variables, constants,
parameters, results, receivers, fields, types, type parameters, functions, labels, the names given to
imports and the name a type switch binds. A name is split into its words (rowIndex: row, Index;
HTTPServer: HTTP, Server; userIDs: user, IDs), and each word must have three letters or more, digits not
counted, and not be an abbreviation, in the singular or the plural (seg, segs, cfg). A few short words are
kept: alone, those of D11; inside a longer name, also the small words of English. An initialism is a word,
written as Go writes it: in capitals, or in small letters where a name or a part of it starts (ID, userID,
idList, Example_url; URLs; ETag), and refused in mixed case (userId, Urls). One of two letters stands
alone only as "id" (D14).`

// nameRule is what the rule reads a word against, each table keyed in lower case.
type nameRule struct {
	// keptNames are the short words that may stand as a name alone
	keptNames map[string]bool
	// keptWords are the short words that may stand inside a longer name
	keptWords map[string]bool
	// abbreviations are words cut short, refused whatever their length
	abbreviations map[string]bool
	// englishWords are the words of English an abbreviation's plural would take for one of it
	englishWords map[string]bool
	// spellings are the initialisms and the names of things, each as Go and its owner write it
	spellings map[string]string
	// mixedSpellings are the spellings with small letters in them, which a split by capitals would break
	mixedSpellings []string
}

// verdict is what the rule finds of one word of a name.
type verdict int

const (
	// whole is a word that says what it means
	whole verdict = iota
	// cut is a word too short, or an abbreviation
	cut
	// mixedCase is an initialism written neither as Go writes it nor in small letters at the start
	mixedCase
)

func wholeNamesAnalyzer() *analysis.Analyzer {
	rule := nameRule{
		keptNames:      setOf(keptShortNames()),
		keptWords:      setOf(slices.Concat(keptShortNames(), smallWords())),
		abbreviations:  setOf(refusedAbbreviations()),
		englishWords:   setOf(wordsLikeAbbreviations()),
		spellings:      spellingsByLowerCase(slices.Concat(goInitialisms(), projectInitialisms())),
		mixedSpellings: mixedSpellingsOf(projectInitialisms()),
	}
	return &analysis.Analyzer{
		Name: "wholenames",
		Doc:  wholeNamesDescription,
		Run:  func(pass *analysis.Pass) (any, error) { return runWholeNames(pass, rule) },
	}
}

// keptShortNames are the names shorter than a word that are kept (docs/FOUNDATION.md §6.1, D11). "err" and
// "ctx" stand in for "error" and "context", which the language has taken.
func keptShortNames() []string {
	return []string{"err", "ctx", "id", "ok", "to", "at", "of", "on", "by", "is", "as"}
}

// smallWords are the short words of English that read as words inside a longer name: `oneOfAKind`, `inOrder`.
func smallWords() []string {
	return []string{"a", "an", "in", "or", "if", "no", "up", "do", "go"}
}

// goInitialisms are the initialisms staticcheck's ST1003 knows by default, copied whole (its check is off, as
// this rule holds names whole): golangci-lint v2.15.0 carries them.
func goInitialisms() []string {
	return []string{
		"ACL", "AMQP", "API", "ASCII", "CPU", "CSS", "DB", "DNS", "EOF", "GID", "GUID", "HTML", "HTTP", "HTTPS",
		"ID", "IP", "JSON", "QPS", "RAM", "RPC", "RTP", "SIP", "SLA", "SMTP", "SQL", "SSH", "TCP", "TLS", "TS",
		"TTL", "UDP", "UI", "UID", "URI", "URL", "UTF8", "UUID", "VM", "XML", "XMPP", "XSRF", "XSS",
	}
}

// projectInitialisms are this project's own: the initialisms of its trade, the names the standard library gives
// a hash, a certificate format and a signature scheme (crypto/md5, crypto/x509, crypto/ed25519), and the names of things as their owners spell them, mixed
// case and all (D14).
func projectInitialisms() []string {
	return []string{
		"CLI", "CSP", "IO", "MD5", "OS", "OSC", "PDF", "PNG", "PTY", "SVG", "TUI", "X509",
		"Ed25519", "ETag", "IPv4", "IPv6", "OAuth",
	}
}

// wordsLikeAbbreviations are words of English that read as an abbreviation's plural: `lens` is no `len`.
func wordsLikeAbbreviations() []string {
	return []string{"lens", "curs"}
}

// refusedAbbreviations are words cut short that the length of a word lets through, each in the singular, and
// the plural of a kept short name, which no longer stands in for a taken one.
func refusedAbbreviations() []string {
	return []string{
		"errs", "ctxs",
		"addr", "arg", "arr", "attr", "btn", "buf", "calc", "cfg", "cmd", "cnt", "col", "conf", "ctrl", "cur",
		"curr", "dec", "def", "desc", "dict", "dir", "doc", "dst", "elem", "env", "fmt", "fsys", "idx",
		"impl", "lbl", "len", "loc", "lst", "mgr", "msg", "num", "obj", "opt", "para", "param", "pkg", "pos",
		"prev", "proj", "ptr", "rec", "ref", "req", "res", "resp", "seg", "sep", "seq", "src", "srv", "str",
		"svc", "tmp", "tpl", "tmpl", "util", "val",
	}
}

// mixedSpellingsOf are the spellings with a small letter in them.
func mixedSpellingsOf(spellings []string) []string {
	return slices.DeleteFunc(slices.Clone(spellings), func(spelling string) bool { return strings.ToUpper(spelling) == spelling })
}

func setOf(words []string) map[string]bool {
	set := make(map[string]bool, len(words))
	for _, word := range words {
		set[word] = true
	}
	return set
}

func spellingsByLowerCase(spellings []string) map[string]string {
	byLowerCase := make(map[string]string, len(spellings))
	for _, spelling := range spellings {
		byLowerCase[strings.ToLower(spelling)] = spelling
	}
	return byLowerCase
}

func runWholeNames(pass *analysis.Pass, rule nameRule) (any, error) {
	for _, file := range writtenFiles(pass) {
		ast.Inspect(file, func(node ast.Node) bool { return visitForNames(pass, rule, node) })
	}
	return nil, nil
}

func visitForNames(pass *analysis.Pass, rule nameRule, node ast.Node) bool {
	switch typed := node.(type) {
	case *ast.File:
		reportName(pass, rule, typed.Name)
	case *ast.Ident:
		if pass.TypesInfo.Defs[typed] != nil {
			reportName(pass, rule, typed)
		}
	case *ast.TypeSwitchStmt:
		reportSwitchedName(pass, rule, typed)
	}
	return true
}

// reportSwitchedName reads the name a type switch binds, which the type checker records once in each case and
// never as a definition of its own.
func reportSwitchedName(pass *analysis.Pass, rule nameRule, statement *ast.TypeSwitchStmt) {
	assignment, isAssignment := statement.Assign.(*ast.AssignStmt)
	if !isAssignment {
		return
	}
	for _, target := range assignment.Lhs {
		if identifier, isIdentifier := target.(*ast.Ident); isIdentifier {
			reportName(pass, rule, identifier)
		}
	}
}

// reportName reports the first word of a name that is not whole. A name of one word stands alone: it is read
// against the names kept alone. A word inside a longer name is read against the words kept there.
func reportName(pass *analysis.Pass, rule nameRule, identifier *ast.Ident) {
	words := wordsOf(identifier.Name, rule.mixedSpellings)
	alone := len(words) == 1
	for _, word := range words {
		switch judgeWord(rule, word, alone) {
		case whole:
		case cut:
			pass.Reportf(identifier.Pos(), cutWordMessage, word, identifier.Name)
			return
		case mixedCase:
			pass.Reportf(identifier.Pos(), mixedCaseMessage, identifier.Name, spellingOf(rule, word))
			return
		}
	}
}

// judgeWord reads one word: a word of English first; then an abbreviation, in the singular or the plural; then
// an initialism, spelled as Go spells it, and alone only if kept; then a kept word; then its letters.
func judgeWord(rule nameRule, word string, alone bool) verdict {
	lower := strings.ToLower(word)
	singular := strings.TrimSuffix(lower, plural)
	kept := rule.keptWords
	if alone {
		kept = rule.keptNames
	}
	switch {
	case rule.englishWords[lower]:
		return whole
	case rule.abbreviations[lower] || rule.abbreviations[singular]:
		return cut
	case spellingOf(rule, word) != "":
		return judgeInitialism(rule, word, alone)
	case kept[lower]:
		return whole
	case letterCount(word) < wholeWordLetters:
		return cut
	default:
		return whole
	}
}

// judgeInitialism holds an initialism to Go's spelling, or to small letters, which a word of a camel-cased
// name has only where the name or a part of it starts; and one shorter than a word, alone, to the kept names.
func judgeInitialism(rule nameRule, word string, alone bool) verdict {
	singular := strings.ToLower(word)
	if _, isSpelling := rule.spellings[singular]; !isSpelling {
		singular = strings.TrimSuffix(singular, plural)
	}
	if alone && letterCount(singular) < wholeWordLetters && !rule.keptNames[singular] {
		return cut
	}
	if word == spellingOf(rule, word) || word == strings.ToLower(word) {
		return whole
	}
	return mixedCase
}

// letterCount counts the letters of a word, and none of its digits: `p99` has one.
func letterCount(word string) int {
	count := 0
	for _, letter := range word {
		if unicode.IsLetter(letter) {
			count++
		}
	}
	return count
}

// spellingOf is how Go writes an initialism, plural or not; empty for a word that is none.
func spellingOf(rule nameRule, word string) string {
	lower := strings.ToLower(word)
	if spelling, isInitialism := rule.spellings[lower]; isInitialism {
		return spelling
	}
	if spelling, isInitialism := rule.spellings[strings.TrimSuffix(lower, plural)]; isInitialism {
		return spelling + plural
	}
	return ""
}

// wordsOf splits a name into its words: at an underscore; where a capital follows a small letter or a digit;
// before the last capital of a run that a small letter follows (HTTPServer: HTTP, Server), but for the small
// s of a plural (userIDs: user, IDs); and round a mixed spelling, kept whole (ETagValue: ETag, Value). A digit
// stays with the word before it, and is no letter of it.
func wordsOf(name string, mixedSpellings []string) []string {
	words := []string{}
	for part := range strings.SplitSeq(name, underscore) {
		letters := []rune(part)
		for len(letters) > 0 {
			length := wordLength(letters, mixedSpellings)
			words = append(words, string(letters[:length]))
			letters = letters[length:]
		}
	}
	return words
}

// wordLength is how many letters the word at the start of letters holds.
func wordLength(letters []rune, mixedSpellings []string) int {
	for _, spelling := range mixedSpellings {
		if length := spelledLength(letters, spelling); length > 0 {
			return length
		}
	}
	for position := 1; position < len(letters); position++ {
		if startsWord(letters, position) {
			return position
		}
	}
	return len(letters)
}

// spelledLength is how many letters a mixed spelling takes at the start of letters, with the small s of its
// plural (staticETags); 0 where letters do not open with it, or where it runs on into a longer word.
func spelledLength(letters []rune, spelling string) int {
	length := utf8.RuneCountInString(spelling)
	switch {
	case len(letters) < length || string(letters[:length]) != spelling:
		return 0
	case len(letters) == length || !unicode.IsLower(letters[length]):
		return length
	case isPluralEnding(letters, length):
		return length + 1
	default:
		return 0
	}
}

// startsWord tells whether the letter at a position, not the first of its word, begins a new one.
func startsWord(letters []rune, position int) bool {
	if !unicode.IsUpper(letters[position]) {
		return false
	}
	if !unicode.IsUpper(letters[position-1]) {
		return true
	}
	next := position + 1
	return next < len(letters) && unicode.IsLower(letters[next]) && !isPluralEnding(letters, next)
}

// isPluralEnding tells the small s of a plural: the last letter of the word it ends.
func isPluralEnding(letters []rune, position int) bool {
	ending := position + 1
	return string(letters[position]) == plural && (ending == len(letters) || !unicode.IsLower(letters[ending]))
}
