package main

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// Each rule runs on a package of its own under testdata/src, where a `// want` comment names every finding the
// rule must report on its line, and a line without one must report nothing.

func TestNamesShorterThanAWordAreRefused(test *testing.T) {
	analysistest.Run(test, analysistest.TestData(), wholeNamesAnalyzer(), "names", "xy")
}

func TestHelpersInsideFunctionsAreRefused(test *testing.T) {
	analysistest.Run(test, analysistest.TestData(), noNestedHelpersAnalyzer(), "helpers")
}

func TestArraysOfWordsAreRefused(test *testing.T) {
	analysistest.Run(test, analysistest.TestData(), noRecordArraysAnalyzer(), "records")
}

func TestMarkupInStringsIsRefused(test *testing.T) {
	analysistest.Run(test, analysistest.TestData(), noMarkupAnalyzer(), "markup")
}

func TestColoursInStringsAreRefused(test *testing.T) {
	analysistest.Run(test, analysistest.TestData(), noColoursAnalyzer(), "colours")
}

func TestNumbersOutsideConstantsAreRefused(test *testing.T) {
	analysistest.Run(test, analysistest.TestData(), noNumbersAnalyzer(), "numbers")
}
