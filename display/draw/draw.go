package draw

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

var (
	screen     [30][40]canvasPixel // the processor virtual screen of characters
	charMIF    [128][8]byte        // the binary representation of characters: an 8x8 bitfield for each ascii character
	viewport   *canvas.Raster      // the viewport to display the pixels from the screen
	icmcColors = []color.RGBA{
		{0xff, 0xff, 0xff, 0},
		{0xa5, 0x2a, 0x2a, 0},
		{0x00, 0xff, 0x00, 0},
		{0x6b, 0x8e, 0x23, 0},
		{0x23, 0x23, 0x8e, 0},
		{0x87, 0x1f, 0x78, 0},
		{0x00, 0x80, 0x80, 0},
		{0xe6, 0xe8, 0xfa, 0},
		{0xbe, 0xbe, 0xbe, 0},
		{0xff, 0x00, 0x00, 0},
		{0x32, 0xcd, 0x32, 0},
		{0xff, 0xff, 0x00, 0},
		{0x00, 0x00, 0xff, 0},
		{0xff, 0x1c, 0xae, 0},
		{0x7a, 0xdb, 0x93, 0},
		{0x20, 0x20, 0x20, 0},
		{0x00, 0x00, 0x00, 0},
	} // all the colors defined by the ICMC architecture
)

// canvasPixel represents a single pixel for the virtual character screen for
// the processor.
type canvasPixel struct {
	col  color.Color
	char byte
}

// drawPixel paints a single pixel at an x and y positions of a screen of size
// w and h. It returns the color that the pixel at that position should be.
//
// Important: the underlying color struct must be the same across all pixels,
// if it isn't fyne might make wrong assumptions about the pixel color data
// type.
func drawPixel(x, y, w, h int) color.Color {
	// if the screen is not in the correct aspect ratio (40x30),
	// resize for our calculations.
	if h*40 < w*30 {
		w = h / 30 * 40
	} else {
		h = w / 40 * 30
	}

	// resize the x and y values for the nearest character pixel index possible
	// (where 0 is the first pixel of the first character,
	// 1 is the second pixel of the first character,
	// 8 is the first pixel of the second character, and so on)
	roundy := y / (h / (30 * 8))
	roundx := x / (w / (40 * 8))

	if roundy/8 >= 30 || roundx/8 >= 40 {
		// if we went past the boundaries of the virtual screen,
		// draw a transparent non background.
		return color.RGBA{0, 0, 0, 0}
	}

	// get the character drawn at that position
	pixel := screen[roundy/8][roundx/8]

	// see if the pixel should be colored or not as defined by the MIF char
	// mapping, meaning: get the byte/scanline for that character based on roundy
	// and if the bit at a position based on roundx is set,
	// draw it in the color speficied.
	bit := charMIF[pixel.char][roundy%8] & (1 << (7 - (roundx % 8)))
	if bit != 0 {
		return pixel.col
	}

	// otherwise, use the background default black color.
	return color.RGBA{0, 0, 0, 0xff}
}

// Reset resets the viewport and makes all characters in the virtual screen be
// '\0'.
func Reset() {
	for i := range screen {
		for j := range screen[i] {
			screen[i][j] = canvasPixel{color.RGBA{0, 0, 0, 0xff}, byte(0)}
		}
	}
	viewport.Refresh()
}

// Refresh refreshes the viewport, drawing the whole viewport again based on
// (hopefully changed) data at the virtual screen.
func Refresh() {
	viewport.Refresh()
}

// MakeViewPort creates a new viewport for the simulator.
func MakeViewPort() *canvas.Raster {
	viewport = canvas.NewRasterWithPixels(drawPixel)
	viewport.SetMinSize(fyne.NewSize(40*8*2, 30*8*2))
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

// toICMCColor gets, based on the high byte of a character to be drawn, it's
// color in RGB. This is a convention based loosely on DOS colors, defined
// previously by the designers of the ICMC architecture.
func toICMCColor(colbyte byte) color.Color {
	if colbyte > 16 {
		return color.RGBA{0, 0, 0, 0}
	}
	return icmcColors[colbyte]
}

// FyneOutChar implements the outchar instruction for the simulator:
// bounds check the position and character being drawn, and write them to the
// virtual screen buffer with the right color.
func FyneOutChar(c, pos uint16) error {
	if pos >= 40*30 {
		return fmt.Errorf("invalid position to draw on")
	}
	if byte(c) > 127 || byte(c<<8) > 16 {
		return fmt.Errorf("invalid character to print")
	}

	screen[pos/40][pos%40].char = byte(c)
	screen[pos/40][pos%40].col = toICMCColor(byte(c >> 8))

	Refresh()
	return nil
}
