package src

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Camera represents an advanced 2D camera that follows the player with smooth interpolation.
type Camera struct {
	X, Y      float64 // Current camera position (top-left corner)
	TargetX   float64 // Target camera position X
	TargetY   float64 // Target camera position Y
	ViewportW float64 // Width of the viewport
	ViewportH float64 // Height of the viewport
	WorldW    float64 // Width of the world (0 = infinite width)
	WorldH    float64 // Height of the world

	// Camera behavior settings
	FollowSpeed    float64 // Speed of camera interpolation (0-1, higher = more responsive)
	LookAhead      float64 // How far ahead of the player to look based on velocity
	DeadZone       float64 // Dead zone around player center where camera doesn't move
	VerticalOffset float64 // Vertical offset from player center (negative = look down)
}

// NewCamera creates a new camera with the given viewport and world size.
func NewCamera(viewportW, viewportH, worldW, worldH float64) *Camera {
	return &Camera{
		ViewportW:      viewportW,
		ViewportH:      viewportH,
		WorldW:         worldW,
		WorldH:         worldH,
		FollowSpeed:    8.0,   // Responsive but smooth
		LookAhead:      0.3,   // Moderate look-ahead
		DeadZone:       20.0,  // Small dead zone
		VerticalOffset: -30.0, // Look slightly down to see ground better
	}
}

// Update the camera with delta time for smooth interpolation
func (c *Camera) Update(deltaTime float64) {
	// Smooth camera position interpolation
	dx := c.TargetX - c.X
	dy := c.TargetY - c.Y

	// Only move camera if outside dead zone or if distance is significant
	if math.Abs(dx) > c.DeadZone || math.Abs(dy) > c.DeadZone ||
		math.Abs(dx) > 1.0 || math.Abs(dy) > 1.0 {
		c.X += dx * c.FollowSpeed * deltaTime
		c.Y += dy * c.FollowSpeed * deltaTime
	}
}

// Follow sets the target position for the camera to follow the player with look-ahead.
func (c *Camera) Follow(playerX, playerY, velocityX, velocityY float64) {
	// Calculate look-ahead based on player velocity
	lookAheadX := velocityX * c.LookAhead
	lookAheadY := velocityY * c.LookAhead * 0.3 // Less vertical look-ahead

	// Calculate desired camera center position
	halfW := c.ViewportW / 2
	halfH := c.ViewportH / 2

	// Set target position (top-left corner of camera view)
	c.TargetX = playerX + lookAheadX - halfW
	c.TargetY = playerY + lookAheadY - halfH + c.VerticalOffset

	// Clamp to world bounds if world bounds are set
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

// GetView returns the camera's current view rectangle
func (c *Camera) GetView() (x, y, w, h float64) {
	return c.X, c.Y, c.ViewportW, c.ViewportH
}

// GetTransform returns a GeoM transform matrix for drawing with camera offset
func (c *Camera) GetTransform() *ebiten.GeoM {
	var transform ebiten.GeoM

	// Apply camera offset
	transform.Translate(-c.X, -c.Y)

	return &transform
}

// WorldToScreen converts world coordinates to screen coordinates
func (c *Camera) WorldToScreen(worldX, worldY float64) (screenX, screenY float64) {
	screenX = worldX - c.X
	screenY = worldY - c.Y
	return screenX, screenY
}

// ScreenToWorld converts screen coordinates to world coordinates
func (c *Camera) ScreenToWorld(screenX, screenY float64) (worldX, worldY float64) {
	worldX = screenX + c.X
	worldY = screenY + c.Y
	return worldX, worldY
}

// SetWorldBounds updates the world boundaries
func (c *Camera) SetWorldBounds(width, height float64) {
	c.WorldW = width
	c.WorldH = height
}

// SetViewport updates the viewport size
func (c *Camera) SetViewport(width, height float64) {
	c.ViewportW = width
	c.ViewportH = height
}

// GetCameraSettings returns current camera behavior settings for debugging
func (c *Camera) GetCameraSettings() (followSpeed, lookAhead, deadZone, verticalOffset float64) {
	return c.FollowSpeed, c.LookAhead, c.DeadZone, c.VerticalOffset
}
