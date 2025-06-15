package assets

import (
	"embed"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/temidaradev/esset/v2"
)

//go:embed *
var assets embed.FS

//go:embed font/PublicPixel.ttf
var Font []byte
var FontFaceS text.Face
var FontFaceM text.Face

var (
	BgZ1         = esset.GetAsset(assets, "images/start/bg_z-1.png")
	BgZ2         = esset.GetAsset(assets, "images/start/bg_z-2.png")
	BgZ3         = esset.GetAsset(assets, "images/start/bg_z-3.png")
	MountainsZ4  = esset.GetAsset(assets, "images/start/mtnz-4.png")
	GradientZ6   = esset.GetAsset(assets, "images/start/Gradientz-6.png")
	MiddleGround = esset.GetAsset(assets, "images/start/middlegroundz0.png")
	Foreground   = esset.GetAsset(assets, "images/start/middleplusz1.png")
)
