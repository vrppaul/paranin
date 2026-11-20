package main

import (
	"image/color"
	"math"
)

// xterm's 256 colours past its first sixteen: a cube of six levels for each of red, green and blue, from index
// 16, its levels nothing and then 95 rising by 40; then a ramp of greys, from index 232, starting at 8 and rising
// by 10
const (
	firstCubeIndex  = 16
	firstGreyIndex  = 232
	lastIndex       = 255
	cubeSide        = 6
	cubeLevelOffset = 55
	cubeLevelStep   = 40
	firstGreyLevel  = 8
	greyStep        = 10
)

// shade is one colour as each kind of terminal shows it: whole, as the nearest of xterm's 256, and as one of
// sixteen, or the terminal's own.
type shade struct {
	trueColour color.RGBA
	indexed    uint8
	basic      basicValue
}

// themedShade is a colour's shade in each theme.
type themedShade struct {
	light shade
	dark  shade
}

// worked is one colour of the palette, worked out: its name, what it is for, and its shade in each theme.
type worked struct {
	name   string
	about  string
	shades themedShade
}

// paletteOf works out every colour of the tokens, then every blend, in the order the file names them.
func paletteOf(read tokens) []worked {
	palette := make([]worked, 0, len(read.colours)+len(read.blends))
	for _, colour := range read.colours {
		palette = append(palette, worked{
			name:  colour.name,
			about: colour.about,
			shades: themedShade{
				light: shadeOf(colour.values.Light, colour.values.Light, colour.basic),
				dark:  shadeOf(colour.values.Dark, colour.values.Dark, colour.basic),
			},
		})
	}
	for _, blend := range read.blends {
		palette = append(palette, worked{
			name:  blend.name,
			about: blend.about,
			shades: themedShade{
				light: blendShade(read.background.Light, blend.over.Light, blend.strengths.Light, blend.indexedLean, blend.basic),
				dark:  blendShade(read.background.Dark, blend.over.Dark, blend.strengths.Dark, blend.indexedLean, blend.basic),
			},
		})
	}
	return palette
}

// blendShade is a colour laid over another at a strength; for 256 colours it leans toward its colour first.
func blendShade(under, over color.RGBA, strength, lean float64, basic basicValue) shade {
	laid := mixed(under, over, strength)
	return shadeOf(laid, mixed(laid, over, lean), basic)
}

// shadeOf is a colour whole, the nearest of xterm's 256 to the colour a terminal of 256 is given, and a basic one.
func shadeOf(whole, forIndexed color.RGBA, basic basicValue) shade {
	return shade{trueColour: whole, indexed: nearestIndexed(forIndexed), basic: basic}
}

// mixed is one colour laid over another at a strength, each channel rounded.
func mixed(base, over color.RGBA, strength float64) color.RGBA {
	return color.RGBA{
		R: mixedChannel(base.R, over.R, strength),
		G: mixedChannel(base.G, over.G, strength),
		B: mixedChannel(base.B, over.B, strength),
		A: opaque,
	}
}

func mixedChannel(base, over uint8, strength float64) uint8 {
	return uint8(math.Round(float64(base)*(strengthOfWhole-strength) + float64(over)*strength))
}

// nearestIndexed is the index of the colour of xterm's 256, past the sixteen a terminal may set its own way,
// closest to a colour.
func nearestIndexed(colour color.RGBA) uint8 {
	nearest, nearestDistance := uint8(firstCubeIndex), math.MaxInt
	for index := uint8(firstCubeIndex); ; index++ {
		if distance := distanceBetween(colour, indexedColour(index)); distance < nearestDistance {
			nearest, nearestDistance = index, distance
		}
		if index == lastIndex {
			return nearest
		}
	}
}

// indexedColour is the colour of an index of xterm's cube or its ramp of greys.
func indexedColour(index uint8) color.RGBA {
	if index >= firstGreyIndex {
		level := firstGreyLevel + (index-firstGreyIndex)*greyStep
		return color.RGBA{R: level, G: level, B: level, A: opaque}
	}
	place := index - firstCubeIndex
	return color.RGBA{
		R: cubeLevel(place / (cubeSide * cubeSide)),
		G: cubeLevel(place / cubeSide % cubeSide),
		B: cubeLevel(place % cubeSide),
		A: opaque,
	}
}

// cubeLevel is the level of a step along a side of xterm's cube: nothing, then 95 rising by 40.
func cubeLevel(step uint8) uint8 {
	if step == 0 {
		return 0
	}
	return cubeLevelOffset + step*cubeLevelStep
}

// distanceBetween is the square of how far apart two colours lie, channel by channel.
func distanceBetween(first, second color.RGBA) int {
	red, green, blue := int(first.R)-int(second.R), int(first.G)-int(second.G), int(first.B)-int(second.B)
	return red*red + green*green + blue*blue
}
