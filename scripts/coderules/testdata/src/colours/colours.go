package colours

const ground = "#eff1f5" // want `a colour is a design token`

const short = "#fff" // want `a colour is a design token`

const shortWithAlpha = "#fff8" // want `a colour is a design token`

const longWithAlpha = "#1e1e2e80" // want `a colour is a design token`

const hue = "hsl(0 0% 0%)" // want `a colour is a design token`

const perceived = "oklch(0.7 0.1 250)" // want `a colour is a design token`

const (
	whiteness = "hwb(0 0% 0%)"        // want `a colour is a design token`
	lightness = "lab(50% 40 59)"      // want `a colour is a design token`
	chroma    = "lch(50% 70 40)"      // want `a colour is a design token`
	linear    = "oklab(0.6 0.1 -0.1)" // want `a colour is a design token`
)

const shade = "rgba(76, 79, 105, .22)" // want `a colour is a design token`

const mixed = "color-mix(in srgb, var(--accent) 18%, transparent)" // want `a colour is a design token`

// A hash or a word in brackets that names no colour is words.
const (
	issue    = "fixed in #123"
	hashtag  = "#paranin"
	sentence = "rgb is the screen's light"
)
