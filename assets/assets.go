package assets

import (
	_ "embed"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

//go:embed font/PublicPixel.ttf
var Font []byte
var FontFaceS text.Face
var FontFaceM text.Face
