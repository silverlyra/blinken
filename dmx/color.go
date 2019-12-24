package dmx

import (
	"fmt"

	"lyra.codes/blinken/color"
)

const (
	channels = 4

	red   = 0
	green = 1
	blue  = 2
	white = 3
)

// RGBW is one or more RGBW colors as DMX channels.
type RGBW []Channel

func (c RGBW) Len() int {
	return len(c) / channels
}

func (c RGBW) At(index int) color.RGBW {
	return color.RGBW{
		byteToFloat(c[index*channels+red]),
		byteToFloat(c[index*channels+green]),
		byteToFloat(c[index*channels+blue]),
		byteToFloat(c[index*channels+white]),
	}
}

func (c RGBW) Set(index int, d color.RGBW) {
	c[index*channels+red] = floatToByte(d.R)
	c[index*channels+green] = floatToByte(d.G)
	c[index*channels+blue] = floatToByte(d.B)
	c[index*channels+white] = floatToByte(d.W)
}

func (c RGBW) Spread(startIndex int, colors []color.RGBW) {
	for i, color := range colors {
		c.Set(startIndex+i, color)
	}
}

func (c RGBW) Inspect() []string {
	l := c.Len()
	out := make([]string, 0, l)

	for i := 0; i < l; i++ {
		out = append(
			out,
			fmt.Sprintf("(%03d, %03d, %03d, %03d)", c[i*channels+red], c[i*channels+green], c[i*channels+blue], c[i*channels+white]),
		)
	}

	return out
}

func byteToFloat(b Channel) float64 {
	return float64(b) / 255.0
}

func floatToByte(f float64) Channel {
	return Channel(f*255.0 + 0.5)
}
