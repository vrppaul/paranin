package main

import (
	"go/ast"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// How the name of a test file ends
const testFileSuffix = "_test.go"

// writtenFiles are the files of the package a person wrote: every one but those a generator wrote.
func writtenFiles(pass *analysis.Pass) []*ast.File {
	return slices.DeleteFunc(slices.Clone(pass.Files), ast.IsGenerated)
}

// shippedFiles are the written files the program is built from: every written one but the tests.
func shippedFiles(pass *analysis.Pass) []*ast.File {
	return slices.DeleteFunc(writtenFiles(pass), func(file *ast.File) bool { return isTestFile(pass, file) })
}

// isTestFile tells a test file by its name, as the go command does.
func isTestFile(pass *analysis.Pass, file *ast.File) bool {
	return strings.HasSuffix(pass.Fset.File(file.Pos()).Name(), testFileSuffix)
}
