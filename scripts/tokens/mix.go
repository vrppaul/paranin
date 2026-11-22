package main

import "image/color"

// The key a mix's share is written under, as its tag says
const shareKey = "share"

// mixToken is a colour the page mixes from two of the file's, as the file writes it: its name, what it is for, the
// colour it starts from, the one it is mixed toward, how far toward it, from nothing to the whole, and where it must
// read, the least contrast it reads at over the grounds it names, as a colour's.
type mixToken struct {
	Name          string         `yaml:"name"`
	About         string         `yaml:"about"`
	Of            string         `yaml:"of"`
	Toward        string         `yaml:"toward"`
	Share         *float64       `yaml:"share"`
	LeastContrast *contrastToken `yaml:"least_contrast"`
}

// readMix is a mix read: its name, what it is for, its colour in each theme, worked out once, as a blend's is, so the
// stylesheet writes it as any colour, and the least contrast it needs, none where it need not read.
type readMix struct {
	name          string
	about         string
	values        themedValue[color.RGBA]
	leastContrast *leastContrast
}

// readMixToken refuses a mix of a colour the file does not name, mixed beyond the whole, or needing a least contrast
// written wrong; each theme's colour is the one it starts from with the share of the other laid over it.
func readMixToken(token mixToken, colours []readColour, named map[string]bool) (readMix, error) {
	if err := checkToken(token.Name, token.About, named); err != nil {
		return readMix{}, err
	}
	of, err := namedColour(colours, token.Of, token.Name)
	if err != nil {
		return readMix{}, err
	}
	toward, err := namedColour(colours, token.Toward, token.Name)
	if err != nil {
		return readMix{}, err
	}
	share, err := strength(token.Share, token.Name, shareKey)
	if err != nil {
		return readMix{}, err
	}
	least, err := readLeastContrast(token.Name, token.LeastContrast)
	if err != nil {
		return readMix{}, err
	}
	values := themedValue[color.RGBA]{Light: mixed(of.Light, toward.Light, share), Dark: mixed(of.Dark, toward.Dark, share)}
	return readMix{name: token.Name, about: token.About, values: values, leastContrast: least}, nil
}

// contrastedOf are the file's colours and its mixes as their least contrasts are checked, each a ground of the others.
func contrastedOf(colours []readColour, mixes []readMix) []contrasted {
	all := make([]contrasted, 0, len(colours)+len(mixes))
	for _, colour := range colours {
		all = append(all, contrasted{name: colour.name, values: colour.values, leastContrast: colour.leastContrast})
	}
	for _, mix := range mixes {
		all = append(all, contrasted{name: mix.name, values: mix.values, leastContrast: mix.leastContrast})
	}
	return all
}
