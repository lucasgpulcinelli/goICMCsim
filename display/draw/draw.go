package draw

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

var (
	screen   [30][40]canvasPixel
	charMIF  [128][8]byte
	viewport *canvas.Raster
)

type canvasPixel struct {
	col  color.Color
	char byte
}

func drawPixel(x, y, w, h int) color.Color {
	if h*40 < w*30 {
		w = h / 30 * 40
	} else {
		h = w / 40 * 30
	}

	roundy := y / (h / (30 * 8))
	roundx := x / (w / (40 * 8))

	if roundy/8 >= 30 || roundx/8 >= 40 {
		return color.RGBA{0, 0, 0, 0}
	}

	pixel := screen[roundy/8][roundx/8]

	bit := charMIF[pixel.char][roundy%8] & (1 << (7 - (roundx % 8)))
	if bit != 0 {
		return pixel.col
	}

	return color.RGBA{0, 0, 0, 0xff}
}

func Reset() {
	for i := range screen {
		for j := range screen[i] {
			screen[i][j] = canvasPixel{color.RGBA{0, 0, 0, 0xff}, byte(0)}
		}
	}
	viewport.Refresh()
}

func Refresh() {
	viewport.Refresh()
}

func MakeViewPort() *canvas.Raster {
	viewport = canvas.NewRasterWithPixels(drawPixel)
	viewport.SetMinSize(fyne.NewSize(40*8*2, 30*8*2))
	Reset()
	return viewport
}

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

func toICMCColor(colbyte byte) color.Color {
	switch colbyte {
	case 0:
		return color.RGBA{0xff, 0xff, 0xff, 0}
	case 1:
		return color.RGBA{0xa5, 0x2a, 0x2a, 0}
	case 2:
		return color.RGBA{0x0, 0xff, 0x0, 0}
	case 3:
		return color.RGBA{0x6b, 0x8e, 0x23, 0}
	case 4:
		return color.RGBA{0x23, 0x23, 0x8e, 0}
	case 5:
		return color.RGBA{0x87, 0x1f, 0x78, 0}
	case 6:
		return color.RGBA{0x0, 0x80, 0x80, 0}
	case 7:
		return color.RGBA{0xe6, 0xe8, 0xfa, 0}
	case 8:
		return color.RGBA{0xbe, 0xbe, 0xbe, 0}
	case 9:
		return color.RGBA{0xff, 0x0, 0x0, 0}
	case 10:
		return color.RGBA{0x32, 0xcd, 0x32, 0}
	case 11:
		return color.RGBA{0xff, 0xff, 0x0, 0}
	case 12:
		return color.RGBA{0x0, 0x0, 0xff, 0}
	case 13:
		return color.RGBA{0xff, 0x1c, 0xae, 0}
	case 14:
		return color.RGBA{0x7a, 0xdb, 0x93, 0}
	case 15:
		return color.RGBA{0x20, 0x20, 0x20, 0}
	}
	return color.RGBA{0, 0, 0, 0}
}

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
