package color

import (
	"fmt"
	"math"

	"github.com/lucasb-eyer/go-colorful"
)

// HCL represents a color in CIE-L*C*h° space.
type HCL struct {
	H float64
	C float64
	L float64
}

func (c HCL) RGB() RGB {
	d := colorful.Hcl(c.H, c.C, c.L).Clamped()
	return RGB{d.R, d.G, d.B}
}

func (c HCL) HSI() HSI {
	return c.RGB().HSI()
}

func (c HCL) RGBW() RGBW {
	return c.RGB().HSI().RGBW()
}

type RGB struct {
	R float64
	G float64
	B float64
}

func (c RGB) Intensity() float64 {
	return (c.R + c.G + c.B) / 3.0
}

func (c RGB) HSI() HSI {
	min := math.Min(c.R, math.Min(c.G, c.B))
	max := math.Max(c.R, math.Max(c.G, c.B))

	h := 0.0
	s := 0.0
	i := c.Intensity()

	if max != min {
		d := max - min

		switch max {
		case c.R:
			h = 60.0 * (0.0 + (c.G-c.B)/d)
		case c.G:
			h = 60.0 * (2.0 + (c.B-c.R)/d)
		case c.B:
			h = 60.0 * (4.0 + (c.R-c.G)/d)
		}

		if h < 0.0 {
			h += 360.0
		}
	}

	if i > 0.0 {
		s = math.Abs(1.0 - min/i)
	}

	return HSI{math.Mod(h, 360.0), clamp1(s), clamp1(i)}
}

const (
	rad60  = 60.0 * math.Pi / 180.0
	rad120 = 120.0 * math.Pi / 180.0
	rad240 = 240.0 * math.Pi / 180.0
)

type HSI struct {
	H float64
	S float64
	I float64
}

func (c HSI) RGBW() RGBW {
	hRad := math.Pi * c.H / 180.0

	cosH := math.Cos(hRad)
	cos60H := math.Cos(rad60 - hRad)

	a := c.S * c.I / 3.0 * (1.0 + cosH/cos60H)
	b := c.S * c.I / 3.0 * (1.0 + (1.0 - cosH/cos60H))
	w := (1.0 - c.S) * c.I

	switch {
	case hRad < rad120:
		return RGBW{clamp1(3.0 * a), clamp1(3.0 * b), 0.0, w}
	case hRad < rad240:
		return RGBW{0.0, clamp1(3.0 * a), clamp1(3.0 * b), w}
	default:
		return RGBW{clamp1(3.0 * b), 0.0, clamp1(3.0 * a), w}
	}
}

type RGBW struct {
	R float64
	G float64
	B float64
	W float64
}

func (c RGBW) String() string {
	return fmt.Sprintf("(r: %0.3f, g: %0.3f, b: %0.3f, w: %0.3f)", c.R, c.G, c.B, c.W)
}

func clamp(v, min, max float64) float64 {
	return math.Max(min, math.Min(v, max))
}

func clamp1(v float64) float64 {
	return clamp(v, 0.0, 1.0)
}
