package main

import (
	"errors"
	"fmt"
	"image/color"
	"math"
)

// How WCAG 2 works out the light a colour gives off: below a share of its channel's whole the light is the share
// over a slope, above it the share raised along a curve; then the channels weighed as the eye weighs them
const (
	linearLimit   = 0.04045
	linearSlope   = 12.92
	curveOffset   = 0.055
	curveScale    = 1.055
	curveExponent = 2.4
	redWeight     = 0.2126
	greenWeight   = 0.7152
	blueWeight    = 0.0722
	// the light a screen gives off at black, added to both sides of a contrast
	flare = 0.05
)

// No contrast at all, and the most two colours have, black on white
const (
	noContrast   = 1
	mostContrast = 21
)

// The faults of a least contrast written wrong
var (
	errUnknownGround = errors.New("a least contrast over a ground the file does not name")
	errNoGround      = errors.New("a least contrast over no ground")
	errRatioBeyond   = errors.New("a least contrast no two colours have")
)

// contrastToken is the least contrast a colour needs as the file writes it: the ratio, and the grounds it needs it
// over, each the page's own colour or a colour of the file.
type contrastToken struct {
	Ratio *float64 `yaml:"ratio"`
	Over  []string `yaml:"over"`
}

// The key a least contrast's ratio is written under, as its tag says
const ratioKey = "ratio"

// leastContrast is the least contrast a colour needs, read: the ratio, and the names of its grounds.
type leastContrast struct {
	ratio float64
	over  []string
}

// underContrast is a colour that reads under its least contrast over a ground in a theme.
type underContrast struct {
	colour string
	ground string
	theme  string
	reads  float64
	needs  float64
}

func (fault underContrast) Error() string {
	return fmt.Sprintf("%v: %q reads %.2f:1 over %q in the %s theme, under the %v:1 it needs",
		errTokens, fault.colour, fault.reads, fault.ground, fault.theme, fault.needs)
}

func (fault underContrast) Unwrap() error {
	return errTokens
}

// readLeastContrast reads the least contrast a colour needs, none where the file writes none: a ratio no two colours
// have, none or beyond black on white, and no ground are each a fault.
func readLeastContrast(colour string, written *contrastToken) (*leastContrast, error) {
	if written == nil {
		return nil, nil
	}
	ratio, err := required(written.Ratio, colour, ratioKey)
	if err != nil {
		return nil, err
	}
	if ratio <= noContrast || ratio > mostContrast {
		return nil, fmt.Errorf("%w: %w: %q needs %v:1, outside %d:1 to %d:1",
			errTokens, errRatioBeyond, colour, ratio, noContrast, mostContrast)
	}
	if len(written.Over) == 0 {
		return nil, fmt.Errorf("%w: %w: %q", errTokens, errNoGround, colour)
	}
	return &leastContrast{ratio: ratio, over: written.Over}, nil
}

// contrasted is a colour as its least contrast is checked, a colour of the file's or a mix of two: its name, its value
// in each theme, and the least contrast it needs, none where it need not read. Each is a ground another may name.
type contrasted struct {
	name          string
	values        themedValue[color.RGBA]
	leastContrast *leastContrast
}

// checkContrasts refuses a colour that reads under its least contrast over a ground it names, in either theme, and
// a ground that is neither the page's own colour nor a colour of the file. Only the page's colours are checked: the
// TUI's ground is the terminal's own, which nothing here can measure.
func checkContrasts(colours []contrasted, background themedValue[color.RGBA]) error {
	grounds := map[string]themedValue[color.RGBA]{backgroundName: background}
	for _, colour := range colours {
		grounds[colour.name] = colour.values
	}
	faults := []error{}
	for _, colour := range colours {
		if colour.leastContrast == nil {
			continue
		}
		for _, groundName := range colour.leastContrast.over {
			ground, known := grounds[groundName]
			if !known {
				faults = append(faults, fmt.Errorf("%w: %w: %q over %q", errTokens, errUnknownGround, colour.name, groundName))
				continue
			}
			needs := colour.leastContrast.ratio
			faults = append(faults,
				readsOver(colour.name, groundName, lightKey, colour.values.Light, ground.Light, needs),
				readsOver(colour.name, groundName, darkKey, colour.values.Dark, ground.Dark, needs))
		}
	}
	return errors.Join(faults...)
}

// readsOver refuses a colour that reads under its least contrast over a ground in a theme.
func readsOver(colour, groundName, theme string, shade, ground color.RGBA, needs float64) error {
	if reads := contrast(shade, ground); reads < needs {
		return underContrast{colour: colour, ground: groundName, theme: theme, reads: reads, needs: needs}
	}
	return nil
}

// contrast is how far apart two colours read, as WCAG 2 works it out: from 1, none, to 21, black on white.
func contrast(first, second color.RGBA) float64 {
	firstLight, secondLight := luminance(first), luminance(second)
	return (max(firstLight, secondLight) + flare) / (min(firstLight, secondLight) + flare)
}

// luminance is the light a colour gives off as the eye weighs it, from none to the whole.
func luminance(colour color.RGBA) float64 {
	return redWeight*linearLight(colour.R) + greenWeight*linearLight(colour.G) + blueWeight*linearLight(colour.B)
}

// linearLight is the light a channel of sRGB gives off, as a share of its whole.
func linearLight(channel uint8) float64 {
	share := float64(channel) / math.MaxUint8
	if share <= linearLimit {
		return share / linearSlope
	}
	return math.Pow((share+curveOffset)/curveScale, curveExponent)
}
