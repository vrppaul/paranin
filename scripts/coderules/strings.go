package main

import (
	"go/ast"
	"go/token"
	"regexp"
	"strconv"

	"golang.org/x/tools/go/analysis"
)

// runStringRule reports every string of the shipped files a pattern finds in, with the rule's message.
func runStringRule(pass *analysis.Pass, pattern *regexp.Regexp, message string) (any, error) {
	for _, file := range shippedFiles(pass) {
		ast.Inspect(file, func(node ast.Node) bool { return visitForStrings(pass, node, pattern, message) })
	}
	return nil, nil
}

func visitForStrings(pass *analysis.Pass, node ast.Node, pattern *regexp.Regexp, message string) bool {
	literal, isLiteral := node.(*ast.BasicLit)
	if !isLiteral || literal.Kind != token.STRING {
		return true
	}
	text, err := strconv.Unquote(literal.Value)
	if err == nil && pattern.MatchString(text) {
		pass.Report(analysis.Diagnostic{Pos: literal.Pos(), Message: message})
	}
	return true
}
