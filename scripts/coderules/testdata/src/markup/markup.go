package markup

const faint = `<span class="faint">` // want `markup lives in a template or a component`

const closing = "</mark>" // want `markup lives in a template or a component`

const shouting = `<DIV ID="top">` // want `markup lives in a template or a component`

const selfClosing = "<br/>" // want `markup lives in a template or a component`

const spacedClosing = "<br />" // want `markup lives in a template or a component`

// Angle brackets round a bare word are words, not markup, whatever the word; so is markdown's autolink.
const (
	usage      = "paranin preview-link <slug> <source>"
	comparison = "a < b and b > a"
	arrow      = "<- done"
	autolink   = "<https://paran.in/>"
)
