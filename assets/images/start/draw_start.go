package start

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/temidaradev/ebijam25/assets"
)

func DrawStart(screen *ebiten.Image, parallaxOffset float64) {
	opBg3 := &ebiten.DrawImageOptions{}
	opBg3.GeoM.Scale(0.8, 0.8)
	opBg3.GeoM.Translate(-parallaxOffset*0.1, 0)
	screen.DrawImage(assets.BgZ3, opBg3)

	opBg2 := &ebiten.DrawImageOptions{}
	opBg2.GeoM.Scale(0.8, 0.8)
	opBg2.GeoM.Translate(-parallaxOffset*0.2, 0)
	screen.DrawImage(assets.BgZ2, opBg2)

	opBg1 := &ebiten.DrawImageOptions{}
	opBg1.GeoM.Scale(0.8, 0.8)
	opBg1.GeoM.Translate(-parallaxOffset*0.3, 0)
	screen.DrawImage(assets.BgZ1, opBg1)

	opMtn := &ebiten.DrawImageOptions{}
	opMtn.GeoM.Scale(0.8, 0.8)
	opMtn.GeoM.Translate(-parallaxOffset*0.5, 0)
	screen.DrawImage(assets.MountainsZ4, opMtn)

	opGradient := &ebiten.DrawImageOptions{}
	opGradient.GeoM.Scale(0.8, 0.8)
	screen.DrawImage(assets.GradientZ6, opGradient)

	opMidGround := &ebiten.DrawImageOptions{}
	opMidGround.GeoM.Scale(0.8, 0.8)
	opMidGround.GeoM.Translate(-parallaxOffset*0.7, 0)
	screen.DrawImage(assets.MiddleGround, opMidGround)

	opForeground := &ebiten.DrawImageOptions{}
	opForeground.GeoM.Scale(0.8, 0.8)
	opForeground.GeoM.Translate(-parallaxOffset*1.0, 0)
	screen.DrawImage(assets.Foreground, opForeground)
}
