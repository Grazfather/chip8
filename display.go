package chip8

const (
	SCREEN_WIDTH  = 64
	SCREEN_HEIGHT = 32
)

func ClearPixel(x, y int, i WriteableImage) {
	i.Set(x, y, 0)
}

type WriteableImage interface {
	At(x, y int) byte
	Set(x, y int, color byte)
	Toggle(x, y int, color byte) (collision bool)
}

type IterableImage interface {
	WriteableImage
	OnEachPixel(func(x, y int, i WriteableImage))
}

type myScreen struct {
	buffer [SCREEN_WIDTH * SCREEN_HEIGHT]byte
}

func (i *myScreen) At(x, y int) byte {
	return i.buffer[y*SCREEN_WIDTH+x]
}

func (i *myScreen) Set(x, y int, color byte) {
	i.buffer[y*SCREEN_WIDTH+x] = color
}

// Toggle will toggle a pixel if color is not zero and return true if the pixel
// was already set, otherwise do nothing.
func (i *myScreen) Toggle(x, y int, color byte) bool {
	// Loop around
	for ; x >= 64; x -= 64 {
	}
	for ; x < 0; x += 64 {
	}
	for ; y >= 32; y -= 32 {
	}
	for ; y < 0; y += 32 {
	}
	a := y*SCREEN_WIDTH + x
	c := i.buffer[a]
	i.buffer[a] ^= color
	// We only check for collisions against set pixels
	if c != 0 && color != 0 {
		return true
	}
	return false
}

func (i *myScreen) OnEachPixel(cb func(x, y int, i WriteableImage)) {
	for y := 0; y < SCREEN_HEIGHT; y++ {
		for x := 0; x < SCREEN_WIDTH; x++ {
			cb(x, y, i)
		}
	}
}

type Renderer interface {
	Render(IterableImage)
}

type NullDisplay struct{}

func (d *NullDisplay) Render(i IterableImage) {}
