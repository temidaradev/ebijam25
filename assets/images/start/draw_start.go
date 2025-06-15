package start

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/temidaradev/ebijam25/assets"
)

func DrawStart(screen *ebiten.Image, parallaxOffset float64) {
	screenWidth := float64(screen.Bounds().Dx())
	screenHeight := float64(screen.Bounds().Dy())

	// Layer config: scaleMode is "height" or "width"
	layers := []struct {
		img       *ebiten.Image
		parallax  float64
		align     string // "top", "bottom", or "center"
		voffset   float64
		scaleMode string // "height" or "width"
	}{
		{assets.GradientZ6, 0.0, "top", 0, "height"},        // Sky
		{assets.BgZ3, 0.03, "top", 0, "height"},             // Distant mountains
		{assets.BgZ2, 0.08, "top", 0, "height"},             // Mid mountains
		{assets.BgZ1, 0.15, "bottom", 0, "width"},           // Forest far
		{assets.MountainsZ4, 0.22, "bottom", 0, "width"},    // Forest mid
		{assets.MiddleGround, 0.35, "bottom", -10, "width"}, // Ground
		{assets.Foreground, 0.5, "bottom", -20, "width"},    // Foreground
	}

	for _, l := range layers {
		if l.scaleMode == "height" {
			drawLayerByHeight(screen, l.img, screenWidth, screenHeight, parallaxOffset*l.parallax, l.align, l.voffset)
		} else {
			drawLayerByWidth(screen, l.img, screenWidth, screenHeight, parallaxOffset*l.parallax, l.align, l.voffset)
		}
	}
}

// Draws a layer, scaling to fit height, aligning vertically, and tiling horizontally for parallax
func drawLayerByHeight(screen *ebiten.Image, img *ebiten.Image, screenWidth, screenHeight float64, offsetX float64, align string, voffset float64) {
	if img == nil {
		return
	}

	imgWidth := float64(img.Bounds().Dx())
	imgHeight := float64(img.Bounds().Dy())

	// Scale to fit screen height
	scale := screenHeight / imgHeight

	// Clamp scale for ground/foreground so they don't get too big
	if scale > 2.0 {
		scale = 2.0
	}
	if scale < 0.5 {
		scale = 0.5
	}

	scaledWidth := imgWidth * scale
	scaledHeight := imgHeight * scale

	// Horizontal parallax tiling
	x := math.Mod(offsetX, scaledWidth)
	if x > 0 {
		x -= scaledWidth
	}

	// Vertical alignment
	var y float64
	switch align {
	case "top":
		y = 0 + voffset
	case "bottom":
		y = screenHeight - scaledHeight + voffset
	default:
		y = (screenHeight-scaledHeight)/2 + voffset
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	for tx := x; tx < screenWidth; tx += scaledWidth {
		op2 := *op
		op2.GeoM.Translate(tx, y)
		screen.DrawImage(img, &op2)
	}
}

// Draws a layer, scaling to fit width, aligning vertically, and tiling horizontally for parallax
func drawLayerByWidth(screen *ebiten.Image, img *ebiten.Image, screenWidth, screenHeight float64, offsetX float64, align string, voffset float64) {
	if img == nil {
		return
	}

	imgWidth := float64(img.Bounds().Dx())
	imgHeight := float64(img.Bounds().Dy())

	// Scale to fit screen width
	scale := screenWidth / imgWidth

	// Clamp scale for ground/foreground so they don't get too big
	if scale > 2.0 {
		scale = 2.0
	}
	if scale < 0.5 {
		scale = 0.5
	}

	scaledWidth := imgWidth * scale
	scaledHeight := imgHeight * scale

	// Horizontal parallax tiling
	x := math.Mod(offsetX, scaledWidth)
	if x > 0 {
		x -= scaledWidth
	}

	// Vertical alignment
	var y float64
	switch align {
	case "top":
		y = 0 + voffset
	case "bottom":
		y = screenHeight - scaledHeight + voffset
	default:
		y = (screenHeight-scaledHeight)/2 + voffset
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	for tx := x; tx < screenWidth; tx += scaledWidth {
		op2 := *op
		op2.GeoM.Translate(tx, y)
		screen.DrawImage(img, &op2)
	}
}
