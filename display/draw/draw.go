package draw

import (
	"fmt"
	"image"
	"image/color"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

const (
	sw  = 40 // screen width
	sh  = 30 // screen height
	chs = 8  // character pixel size, both horizontal and vertical
)

var (
	charactersDrawn [sh][sw]uint16  // the characters previously drawn. Used when changing charmaps during runtime
	screen          *image.Paletted // the actual image with the simulator output characters
	charMIF         [128][8]byte    // the binary representation of characters: an 8x8 bitfield for each ascii character
	viewport        *canvas.Image   // the fyne component to display screen
	shouldDraw      atomic.Value    // an atomic variable to ease the draw thread but keep it from missing updates
	icmcColors      = []color.Color{
		color.RGBA{0xff, 0xff, 0xff, 0xff},
		color.RGBA{0xa5, 0x2a, 0x2a, 0xff},
		color.RGBA{0x00, 0xff, 0x00, 0xff},
		color.RGBA{0x6b, 0x8e, 0x23, 0xff},
		color.RGBA{0x23, 0x23, 0x8e, 0xff},
		color.RGBA{0x87, 0x1f, 0x78, 0xff},
		color.RGBA{0x00, 0x80, 0x80, 0xff},
		color.RGBA{0xe6, 0xe8, 0xfa, 0xff},
		color.RGBA{0xbe, 0xbe, 0xbe, 0xff},
		color.RGBA{0xff, 0x00, 0x00, 0xff},
		color.RGBA{0x32, 0xcd, 0x32, 0xff},
		color.RGBA{0xff, 0xff, 0x00, 0xff},
		color.RGBA{0x00, 0x00, 0xff, 0xff},
		color.RGBA{0xff, 0x1c, 0xae, 0xff},
		color.RGBA{0x7a, 0xdb, 0x93, 0xff},
		color.RGBA{0x20, 0x20, 0x20, 0xff},
		color.RGBA{0x00, 0x00, 0x00, 0xff},
	} // all the colors defined by the ICMC architecture
)

// Reset resets the viewport and makes all characters in the virtual screen be
// '\0'.
func Reset() {
	for i := 0; i < sh; i++ {
		for j := 0; j < sw; j++ {
			charactersDrawn[i][j] = 16 << 8
		}
	}
	RedrawScreen()
}

// RedrawScreen redraws the entire screen with new characters.
func RedrawScreen() {
	for i := 0; i < sh; i++ {
		for j := 0; j < sw; j++ {
			updateChar(j, i, charactersDrawn[i][j])
		}
	}
	viewport.Refresh()
}

// MakeViewPort creates a new viewport for the simulator.
func MakeViewPort() *canvas.Image {
	screen = image.NewPaletted(
		image.Rect(0, 0, sw*chs, sh*chs), icmcColors,
	)

	viewport = canvas.NewImageFromImage(screen)
	viewport.FillMode = canvas.ImageFillContain
	viewport.ScaleMode = canvas.ImageScalePixels

	viewport.SetMinSize(fyne.NewSize(sw*10, sh*10))

	// this gofunc remais forever, called the "draw thread" because, when
	// FyneOutChar runs, it sets an atomic variable for the screen to be redrawn.
	// This is better than running the update on FyneOutChar because in most
	// cases outchar is the bottleneck for the simulator.
	go func() {
		for {
			if shouldDraw.Swap(0) != 0 {
				RedrawScreen()
			}
			// runs at 60 fps
			time.Sleep(17 * time.Millisecond)
		}
	}()

	Reset()
	return viewport
}

// SetCharData sets the mapping for every ascii character for the simulation.
// The data array must have exactaly the size for all scanlines for every ascii
// character (meaning 1024 bytes). Usually, the data is the output of a MIF
// file parsing.
func SetCharData(data []byte) error {
	if len(data) != 128*8 {
		return fmt.Errorf("invalid data size for character data")
	}

	for i := range charMIF {
		for j := range charMIF[i] {
			charMIF[i][j] = data[i*len(charMIF[i])+j]
		}
	}
	return nil
}

// UpdateChar sets the character at position x, y (where x <= 40, y <= 30)
// using the char MIF mapping to c, where the color is at it's higher byte.
func updateChar(x, y int, c uint16) {
	for i := 0; i < chs; i++ {
		scanline := charMIF[uint8(c)][i]
		for j := 0; j < chs; j++ {

			bit := scanline & (1 << (7 - j))

			colorId := uint8(16)
			if bit != 0 {
				colorId = uint8(c >> 8)
			}

			// the actual pixel positions:
			// x and y have character granularity, so each increase goes past
			// chs pixels;
			// i and j have virtual pixel granularity

			px := x*chs + j
			py := y*chs + i
			screen.SetColorIndex(px, py, colorId)
		}
	}
}

// FyneOutChar implements the outchar instruction for the simulator:
// bounds check the position and character being drawn, and write them to the
// virtual screen.
func FyneOutChar(c, pos uint16) error {
	if pos >= sh*sw {
		return fmt.Errorf("invalid position to draw on")
	}
	if byte(c) > 127 || byte(c>>8) > 16 {
		return fmt.Errorf("invalid character to print")
	}

	charactersDrawn[pos/sw][pos%sw] = c
	shouldDraw.Store(1)

	return nil
}
