package dmx

import (
	"fmt"

	"github.com/lucasb-eyer/go-colorful"
)

const (
	channels = 3

	red   = 0
	green = 1
	blue  = 2
)

// RGB is one or more RGB colors as DMX channels.
type RGB []Channel

func (c RGB) Len() int {
	return len(c) / 3
}

func (c RGB) At(index int) colorful.Color {
	return colorful.FastLinearRgb(
		byteToFloat(c[index*channels+red]),
		byteToFloat(c[index*channels+green]),
		byteToFloat(c[index*channels+blue]),
	)
}

func (c RGB) Set(index int, color colorful.Color) {
	r, g, b := color.FastLinearRgb()

	c[index*channels+red] = floatToByte(r)
	c[index*channels+green] = floatToByte(g)
	c[index*channels+blue] = floatToByte(b)
}

func (c RGB) Spread(startIndex int, colors []colorful.Color) {
	for i, color := range colors {
		c.Set(startIndex+i, color)
	}
}

func (c RGB) ToCSS() []string {
	l := c.Len()
	out := make([]string, 0, l)

	for i := 0; i < l; i++ {
		out = append(
			out,
			fmt.Sprintf("#%X%X%X", c[i*channels+red], c[i*channels+green], c[i*channels+blue]),
		)
	}

	return out
}

func byteToFloat(b Channel) float64 {
	return float64(b) / 255.0
}

func floatToByte(f float64) Channel {
	return Channel(f * 255.0)
}
