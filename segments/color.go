package segments

import (
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/diamondburned/cchat/text"
	"github.com/diamondburned/cchat/utils/empty"
	"github.com/lucasb-eyer/go-colorful"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type ColoredSegment struct {
	empty.TextSegment
	strlen  int
	colored Colored
}

var _ text.Segment = (*ColoredSegment)(nil)

func NewColoredSegment(str string, color uint32) ColoredSegment {
	return ColoredSegment{
		strlen:  len(str),
		colored: NewColored(color),
	}
}

func NewRandomColoredSegment(str string) ColoredSegment {
	return ColoredSegment{
		strlen:  len(str),
		colored: NewRandomColored(),
	}
}

func NewColorfulSegment(str string, color colorful.Color) ColoredSegment {
	return ColoredSegment{
		strlen:  len(str),
		colored: NewColorful(color),
	}
}

func (seg ColoredSegment) Bounds() (start, end int) {
	return 0, seg.strlen
}

func (seg ColoredSegment) AsColorer() text.Colorer {
	return seg.colored
}

type Colored uint32

var _ text.Colorer = (*Colored)(nil)

// NewColored makes a new color segment from a string and a 24-bit color.
func NewColored(color uint32) Colored {
	return Colored(color | (0xFF << 24)) // set alpha bits to 0xFF
}

// NewRandomColored returns a random color segment.
func NewRandomColored() Colored {
	return Colored(RandomColor())
}

// NewColorful returns a color segment from the given colorful.Color.
func NewColorful(color colorful.Color) Colored {
	r, g, b := color.RGB255()
	h := (0xFF << 24) + (uint32(r) << 16) + (uint32(g) << 8) + (uint32(b))
	return NewColored(h)
}

func (color Colored) Color() uint32 {
	return uint32(color)
}

var Colors = []uint32{
	0xF5ABBAFF,
	0x5ACFFAFF, // starts here
	0xF5ABBAFF,
	0xFFFFFFFF,
}

var colorIndex uint32 = 0

// RandomColor returns a random 32-bit RGBA color from the known palette.
func RandomColor() uint32 {
	i := atomic.AddUint32(&colorIndex, 1) % uint32(len(Colors))
	return Colors[i]
}
