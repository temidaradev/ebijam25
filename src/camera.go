package src

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type Camera struct {
	X, Y      float64
	TargetX   float64
	TargetY   float64
	ViewportW float64
	ViewportH float64
	WorldW    float64
	WorldH    float64

	FollowSpeed    float64
	LookAhead      float64
	DeadZone       float64
	VerticalOffset float64
}

func NewCamera(viewportW, viewportH, worldW, worldH float64) *Camera {
	return &Camera{
		ViewportW:      viewportW,
		ViewportH:      viewportH,
		WorldW:         worldW,
		WorldH:         worldH,
		FollowSpeed:    8.0,
		LookAhead:      0.3,
		DeadZone:       20.0,
		VerticalOffset: -30.0,
	}
}

func (c *Camera) Update(deltaTime float64) {
	dx := c.TargetX - c.X
	dy := c.TargetY - c.Y

	if math.Abs(dx) > c.DeadZone || math.Abs(dy) > c.DeadZone ||
		math.Abs(dx) > 1.0 || math.Abs(dy) > 1.0 {
		c.X += dx * c.FollowSpeed * deltaTime
		c.Y += dy * c.FollowSpeed * deltaTime
	}
}

func (c *Camera) Follow(playerX, playerY, velocityX, velocityY float64) {
	lookAheadX := velocityX * c.LookAhead
	lookAheadY := velocityY * c.LookAhead * 0.3

	halfW := c.ViewportW / 2
	halfH := c.ViewportH / 2

	c.TargetX = playerX + lookAheadX - halfW
	c.TargetY = playerY + lookAheadY - halfH + c.VerticalOffset

	if c.WorldW > 0 {
		if c.TargetX < 0 {
			c.TargetX = 0
		}
		if c.TargetX > c.WorldW-c.ViewportW {
			c.TargetX = c.WorldW - c.ViewportW
		}
	}

	if c.WorldH > 0 {
		if c.TargetY < 0 {
			c.TargetY = 0
		}
		if c.TargetY > c.WorldH-c.ViewportH {
			c.TargetY = c.WorldH - c.ViewportH
		}
	}
}

func (c *Camera) GetView() (x, y float64) {
	return c.X, c.Y
}

func (c *Camera) GetTransform() *ebiten.GeoM {
	var transform ebiten.GeoM

	transform.Translate(-c.X, -c.Y)

	return &transform
}

func (c *Camera) WorldToScreen(worldX, worldY float64) (screenX, screenY float64) {
	screenX = worldX - c.X
	screenY = worldY - c.Y
	return screenX, screenY
}

func (c *Camera) ScreenToWorld(screenX, screenY float64) (worldX, worldY float64) {
	worldX = screenX + c.X
	worldY = screenY + c.Y
	return worldX, worldY
}

func (c *Camera) SetWorldBounds(width, height float64) {
	c.WorldW = width
	c.WorldH = height
}

func (c *Camera) SetViewport(width, height float64) {
	c.ViewportW = width
	c.ViewportH = height
}

func (c *Camera) GetCameraSettings() (followSpeed, lookAhead, deadZone, verticalOffset float64) {
	return c.FollowSpeed, c.LookAhead, c.DeadZone, c.VerticalOffset
}
