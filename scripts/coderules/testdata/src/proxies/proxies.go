package proxies

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

// The words a fact stands with
const factMark = ": "

func shout(text string) string { // want `shout only passes its call on`
	return strings.ToUpper(text)
}

// title is a text as a title is written.
//
//wrapper:names how a title is written
func title(text string) string {
	return strings.ToUpper(text)
}

// headline is a text as a headline is written, the reason before another directive.
//
//wrapper:names how a headline is written
//go:noinline
func headline(text string) string {
	return strings.ToUpper(text)
}

// A reason must be a directive: with a space after the slashes it is a comment, which a doc prints.
//
// wrapper: says why
func whisper(text string) string { // want `whisper only passes its call on`
	return strings.ToLower(text)
}

func say(text string) { // want `say only passes its call on`
	fmt.Println(text)
}

func held(text string) string { // want `held only passes its call on`
	upper := strings.ToUpper(text)
	return upper
}

func read(path string) ([]byte, error) { // want `read only passes its call on`
	contents, err := os.ReadFile(path)
	return contents, err
}

func loud(text string) string { // want `loud only passes its call on`
	return (strings.ToUpper)(text)
}

func louder(text string) string { // want `louder only passes its call on`
	return (strings.ToUpper(text))
}

func largest(numbers []int) int { // want `largest only passes its call on`
	return slices.Max[[]int](numbers)
}

func inserted(numbers []int, number int) []int { // want `inserted only passes its call on`
	return slices.Insert[[]int, int](numbers, 0, number)
}

//wrapper:
func quiet(text string) string { // want `quiet only passes its call on`
	return strings.ToLower(text)
}

type card struct {
	title string
}

func (shown card) heading() string { // want `heading only passes its call on`
	return strings.TrimSpace(shown.title)
}

// Work done on what is handed is no passing on: a sum, a call inside a call, a value put together, a call made to
// find what to call, a value returned other than the one the call gave.
func exclaimed(text string) string {
	return strings.ToUpper(text) + factMark
}

func doubled(text string, count int) string {
	return strings.Repeat(text, count+count)
}

func trimmedUpper(text string) string {
	return strings.ToUpper(strings.TrimSpace(text))
}

func firstTitle(cards []card) string {
	return strings.ToUpper(slices.Clone(cards)[0].title)
}

func shouter() func(string) string {
	return strings.ToUpper
}

func shoutedBy(text string) string {
	return shouter()(text)
}

func handledBy(handlers []func(string) string, text string) string {
	return handlers[0](text)
}

func kept(text string) string {
	upper := strings.ToUpper(text)
	return text + upper
}

func given(text string) string {
	_ = strings.ToUpper(text)
	return text
}

// A conversion names a type, not a call.
type label string

func labelOf(text string) label {
	return label(text)
}

// A method of a type that implements an interface asking for it is found by its name.
type shape interface {
	Name() string
	Area() int
}

type Card struct {
	Title string
}

func (shown Card) Name() string {
	return strings.TrimSpace(shown.Title)
}

func (shown Card) Area() int {
	return len(shown.Title)
}

func (shown Card) String() string {
	return strings.TrimSpace(shown.Title)
}

func (shown Card) Upper() string { // want `Upper only passes its call on`
	return strings.ToUpper(shown.Title)
}

// A type whose pointer implements an interface lends it its methods of a value receiver too.
type stacker interface {
	Push(text string)
	Size() int
}

type Pile struct {
	Items []string
}

func (held *Pile) Push(text string) {
	held.Items = append(held.Items, text)
}

func (held Pile) Size() int {
	return len(held.Items)
}

type fault struct {
	text string
}

func (failed fault) Error() string {
	return fmt.Sprint(failed.text)
}

// A method of the name alone, of a type that implements no interface asking for it, is not.
type Tag struct {
	Text string
}

func (held Tag) Name() string { // want `Name only passes its call on`
	return strings.TrimSpace(held.Text)
}

func (held Tag) Lower() string { // want `Lower only passes its call on`
	return strings.ToLower(held.Text)
}

// An exported getter of an unexported field is its package's way to it; an unexported one is not.
type Store struct {
	current string
}

func (store Store) Current() string {
	return strings.Clone(store.current)
}

func (store *Store) Latest() string {
	return strings.Clone(store.current)
}

func CurrentOf(store Store) string {
	return strings.Clone(store.current)
}

// Shelf is another name of Store, which a method may be declared on.
type Shelf = Store

func (store Shelf) Oldest() string {
	return strings.Clone(store.current)
}

func (store Store) peek() string { // want `peek only passes its call on`
	return strings.Clone(store.current)
}

// A method of a type kept in its package is reached from nowhere outside, so it is no getter.
type stack struct {
	top string
}

func (held stack) Peek() string { // want `Peek only passes its call on`
	return strings.Clone(held.top)
}
