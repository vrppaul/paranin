package main

import (
	"go/ast"
	"go/constant"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

const numberMessage = "a number with a meaning is a named constant: name what it is"

const noNumbersDescription = `refuse a number written outside a constant

A number other than 0 and 1, the ends of every count and every share, is written once, in a constant whose
name says what it is: a width, a timeout, a position, the length of an array. Anywhere else, in an
assignment, an argument, an index or a type, it is refused. A test may write the numbers it expects.`

func noNumbersAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{Name: "nonumbers", Doc: noNumbersDescription, Run: runNoNumbers}
}

func runNoNumbers(pass *analysis.Pass) (any, error) {
	for _, file := range shippedFiles(pass) {
		ast.Inspect(file, func(node ast.Node) bool { return visitForNumbers(pass, node) })
	}
	return nil, nil
}

// visitForNumbers walks past every declaration of constants, where a number is named.
func visitForNumbers(pass *analysis.Pass, node ast.Node) bool {
	switch typed := node.(type) {
	case *ast.GenDecl:
		return typed.Tok != token.CONST
	case *ast.BasicLit:
		if isNumber(typed) && !isZeroOrOne(typed) {
			pass.Reportf(typed.Pos(), numberMessage)
		}
	}
	return true
}

func isNumber(literal *ast.BasicLit) bool {
	return literal.Kind == token.INT || literal.Kind == token.FLOAT || literal.Kind == token.IMAG
}

func isZeroOrOne(literal *ast.BasicLit) bool {
	value := constant.MakeFromLiteral(literal.Value, literal.Kind, 0)
	return constant.Compare(value, token.EQL, constant.MakeInt64(0)) ||
		constant.Compare(value, token.EQL, constant.MakeInt64(1))
}
