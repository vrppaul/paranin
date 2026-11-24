package content

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/feature/plural"
	"golang.org/x/text/language"
)

// Count is one of the site's phrases that count something, written in each plural form its language gives a whole
// number (CLDR): how many posts a group of the blog's list holds, with that number; the pick that lists every part of a
// series, with the number of parts; and how many of a series' parts are out, with that number and the number of parts.
// The number counted is the phrase's first value, and picks its form.
type Count int

const (
	// The zero count names none
	_ Count = iota
	CountPosts
	CountAllParts
	CountPartsOut
)

// countForms are the keys a catalogue writes each count under, and the values each takes, its number among them.
func countForms() map[Count]phraseForm {
	return map[Count]phraseForm{
		CountPosts:    {key: "posts_count", values: oneValue},
		CountAllParts: {key: "all_parts", values: oneValue},
		CountPartsOut: {key: "parts_out", values: twoValues},
	}
}

// pluralForm is a plural form as a catalogue names it, by CLDR's name for it.
type pluralForm struct {
	name string
	form plural.Form
}

// pluralForms are CLDR's plural forms, in CLDR's order.
func pluralForms() []pluralForm {
	return []pluralForm{
		{name: "zero", form: plural.Zero},
		{name: "one", form: plural.One},
		{name: "two", form: plural.Two},
		{name: "few", form: plural.Few},
		{name: "many", form: plural.Many},
		{name: "other", form: plural.Other},
	}
}

// The numbers a language's forms are read off, from nought: CLDR's rule for a whole number reads the number below a
// thousand, its last digit and its last two, so every form it gives a whole number shows by a thousand. Breton's alone
// reads a million's multiples, which a catalogue in it would have to leave out
const lastNumberRead = 1000

// A whole number has no digits after its point, which CLDR's rule reads as its operands v, w, f and t
const noFraction = 0

// The key a catalogue names the months under, January first, each as a day writes it, 21 Nov 2026, and how many it
// names
const (
	monthsKey  = "months"
	yearMonths = 12
)

// What separates the forms a fault names
const formSeparator = ", "

// Count is a counted phrase as this language writes a number of it, in the form its plural rule gives the number: the
// number in its first place, then each other value in its own.
func (words Words) Count(count Count, number int, values ...string) string {
	return filled(words.counts[count][formOf(words.plurals, number)], slices.Concat([]string{strconv.Itoa(number)}, values))
}

// Month is a month's name as a day in this language writes it.
func (words Words) Month(month time.Month) string {
	return words.months[month-time.January]
}

// formOf is the plural form a language's rule gives a whole number.
//
//wrapper:names a whole number's form, where CLDR's rule asks for the operands of a number with a fraction
func formOf(plurals language.Tag, number int) plural.Form {
	return plural.Cardinal.MatchPlural(plurals, number, noFraction, noFraction, noFraction, noFraction)
}

// pluralRuleOf is a language as CLDR knows it, by its tag, and whether CLDR knows it.
func pluralRuleOf(tag string) (language.Tag, bool) {
	known, err := language.Parse(tag)
	return known, err == nil
}

// wholeForms are the plural forms a language gives a whole number, in CLDR's order.
func wholeForms(plurals language.Tag) []pluralForm {
	given := make([]plural.Form, 0, lastNumberRead+1)
	for number := range lastNumberRead + 1 {
		given = append(given, formOf(plurals, number))
	}
	return slices.DeleteFunc(pluralForms(), func(form pluralForm) bool { return !slices.Contains(given, form.form) })
}

// formNames are plural forms by their names, as a fault lists them.
func formNames(forms []pluralForm) string {
	names := make([]string, 0, len(forms))
	for _, form := range forms {
		names = append(names, form.name)
	}
	return strings.Join(names, formSeparator)
}

// count reads a count a catalogue writes under its key: a text for each form its language gives a whole number, each
// with a place for each of its values, and none for any other form; its texts by their forms.
func (inspection *inspection) count(written writtenWords, counted phraseForm, whole []pluralForm) map[plural.Form]string {
	node, isWritten := written[counted.key]
	texts := map[string]string{}
	switch {
	case !isWritten:
		inspection.add(counted.key, emptyProblem)
		return nil
	case node.Decode(&texts) != nil:
		inspection.add(counted.key, formsProblem, formNames(whole))
		return nil
	}
	read := make(map[plural.Form]string, len(whole))
	for _, given := range whole {
		field := keyPlace(counted.key, given.name)
		read[given.form] = texts[given.name]
		inspection.require(field, read[given.form])
		inspection.requirePlaces(field, counted.values, read[given.form])
	}
	for _, name := range slices.Sorted(maps.Keys(texts)) {
		if !slices.ContainsFunc(whole, func(given pluralForm) bool { return given.name == name }) {
			inspection.add(keyPlace(counted.key, name), strayFormProblem, formNames(whole))
		}
	}
	return read
}

// months reads the months' names a catalogue writes, January first, each a line.
func (inspection *inspection) months(written writtenWords) []string {
	node, isWritten := written[monthsKey]
	var months []string
	switch {
	case !isWritten:
		inspection.add(monthsKey, emptyProblem)
		return nil
	case node.Decode(&months) != nil || len(months) != yearMonths:
		inspection.add(monthsKey, monthsProblem, yearMonths)
		return nil
	}
	for index, month := range months {
		inspection.requireLine(fmt.Sprintf(itemPlace, monthsKey, index), month)
	}
	return months
}
