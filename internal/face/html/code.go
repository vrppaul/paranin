package html

import "github.com/vrppaul/paranin/internal/document"

// The class of each kind of word in code the stylesheet colours (DESIGN.md, Article)
const (
	codeKeywordClass     = "code-keyword"
	codeStringClass      = "code-string"
	codeNumberClass      = "code-number"
	codeFunctionClass    = "code-function"
	codeTypeClass        = "code-type"
	codeBuiltinClass     = "code-builtin"
	codeOperatorClass    = "code-operator"
	codeVariableClass    = "code-variable"
	codeCommentClass     = "code-comment"
	codePunctuationClass = "code-punctuation"
)

// codeClass is the class of a kind of word in code, and whether the kind is coloured at all: a plain word stands in
// the text's colour, as bare text. exhaustive holds the switch to every kind.
func codeClass(kind document.CodeKind) (string, bool) {
	switch kind {
	case document.CodeKeyword:
		return codeKeywordClass, true
	case document.CodeString:
		return codeStringClass, true
	case document.CodeNumber:
		return codeNumberClass, true
	case document.CodeFunction:
		return codeFunctionClass, true
	case document.CodeType:
		return codeTypeClass, true
	case document.CodeBuiltin:
		return codeBuiltinClass, true
	case document.CodeOperator:
		return codeOperatorClass, true
	case document.CodeVariable:
		return codeVariableClass, true
	case document.CodeComment:
		return codeCommentClass, true
	case document.CodePunctuation:
		return codePunctuationClass, true
	case document.CodePlain:
	}
	return "", false
}
