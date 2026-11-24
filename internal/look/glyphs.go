package look

// The terminal's glyphs the design draws in more than one face (DESIGN.md, Iconography): the tick before a thing
// done, a check passed or a command copied; the dash before each point of an entry or a list; the mark before a
// command a reader may run, as a shell's prompt ends in; the three stars a post pauses on, as a book sets them; and the
// arrow before what a version of a post says under its day: an older version's way to the newer, a translation older
// than its original's way to the original
const (
	DoneMark        = "✓"
	PointMark       = "-"
	PromptMark      = "❯"
	PauseMark       = "*   *   *"
	VersionNoteMark = "↑"
)
