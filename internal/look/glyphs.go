package look

// The terminal's glyphs the design draws in more than one face (DESIGN.md, Iconography): the tick before a thing
// done, a check passed or a command copied; the dash before each point of an entry or a list; the mark before a
// command a reader may run, as a shell's prompt ends in; the three stars a post pauses on, as a book sets them; and the
// arrow before the way from an older version of a post to the newer
const (
	DoneMark   = "✓"
	PointMark  = "-"
	PromptMark = "❯"
	PauseMark  = "*   *   *"
	NewerMark  = "↑"
)
