package segments

import (
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/diamondburned/cchat/text"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Colored struct {
	strlen int
	color  uint32
}

var (
	_ text.Colorer = (*Colored)(nil)
	_ text.Segment = (*Colored)(nil)
)

func NewColored(str string, color uint32) Colored {
	return Colored{len(str), color}
}

func NewRandomColored(str string) Colored {
	return Colored{len(str), RandomColor()}
}

func (color Colored) Bounds() (start, end int) {
	return 0, color.strlen - 1
}

func (color Colored) Color() uint32 {
	return color.color
}

// var Colors = []uint32{
// 	0x55cdfc,
// 	0x609ffb,
// 	0x6b78fa,
// 	0x9375f9,
// 	0xc180f8,
// 	0xe88af8,
// 	0xf794e7,
// 	0xf79ecc,
// 	0xf7a8b8,
// }

// func randomColor() uint32 {
// 	return Colors[rand.Intn(len(Colors))]
// }

var Colors = []uint32{
	0xF5ABBA,
	0x5ACFFA, // starts here
	0xF5ABBA,
	0xFFFFFF,
}

var colorIndex uint32 = 0

func RandomColor() uint32 {
	i := atomic.AddUint32(&colorIndex, 1) % uint32(len(Colors))
	return Colors[i]
}
