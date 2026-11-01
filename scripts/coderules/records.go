package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const noRecordArraysDescription = `refuse an array of words standing in for a record

An array of a fixed length whose items are strings or interfaces is a record written by position: a
[2]string of a label and a value says which is which nowhere. A record is a struct with named fields. An
array of numbers is a vector or a digest, and a slice of words is a list: neither is refused. A slice of
slices of words, or a map keyed by position, can stand in for a record too, past what the rule sees.`

const recordMessage = "an array of words stands in for a record: make it a struct with named fields"

func noRecordArraysAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{Name: "norecordarrays", Doc: noRecordArraysDescription, Run: runNoRecordArrays}
}

func runNoRecordArrays(pass *analysis.Pass) (any, error) {
	for _, file := range writtenFiles(pass) {
		ast.Inspect(file, func(node ast.Node) bool { return visitForRecords(pass, node) })
	}
	return nil, nil
}

func visitForRecords(pass *analysis.Pass, node ast.Node) bool {
	array, isArray := node.(*ast.ArrayType)
	if isArray && array.Len != nil && holdsWords(pass.TypesInfo.TypeOf(array.Elt)) {
		pass.Reportf(array.Pos(), recordMessage)
	}
	return true
}

func holdsWords(item types.Type) bool {
	if item == nil {
		return false
	}
	switch underlying := item.Underlying().(type) {
	case *types.Basic:
		return underlying.Info()&types.IsString != 0
	case *types.Interface:
		return true
	default:
		return false
	}
}
