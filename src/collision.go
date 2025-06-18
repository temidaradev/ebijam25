package src

import (
	"github.com/temidaradev/ebijam25/assets"
)

// CollisionSystem handles collision detection and resolution for game objects
type CollisionSystem struct {
	TileMap *assets.TileMap
}

// NewCollisionSystem creates a new collision system
func NewCollisionSystem(tileMap *assets.TileMap) *CollisionSystem {
	return &CollisionSystem{
		TileMap: tileMap,
	}
}

// CollisionBox represents a collision box
type CollisionBox struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// CollisionInfo contains information about a collision
type CollisionInfo struct {
	HasCollision bool
	NewX         float64
	NewY         float64
	HitWall      bool
	HitGround    bool
	HitCeiling   bool
}

// CheckMovement checks if an object can move from one position to another
func (cs *CollisionSystem) CheckMovement(from, to CollisionBox) CollisionInfo {
	info := CollisionInfo{
		HasCollision: false,
		NewX:         to.X,
		NewY:         to.Y,
	}

	if cs.TileMap == nil {
		return info
	}

	// Use the advanced collision system from TileMap
	result := cs.TileMap.CheckMovementAdvanced(from.X, from.Y, to.X, to.Y, to.Width, to.Height)

	info.HasCollision = result.HasCollision
	info.NewX = result.AdjustedX
	info.NewY = result.AdjustedY

	// Determine collision direction
	if result.CollisionX {
		info.HitWall = true
	}
	if result.CollisionY {
		if to.Y > from.Y {
			info.HitGround = true
		} else {
			info.HitCeiling = true
		}
	}

	return info
}

// CheckCollisionAtPoint checks if there's a collision at a specific point
func (cs *CollisionSystem) CheckCollisionAtPoint(box CollisionBox) bool {
	if cs.TileMap == nil {
		return false
	}
	return cs.TileMap.CheckCollision(box.X, box.Y, box.Width, box.Height)
}

// IsOnGround checks if an object is on the ground
func (cs *CollisionSystem) IsOnGround(box CollisionBox) bool {
	if cs.TileMap == nil {
		return false
	}

	// Check multiple positions below the object for more reliable ground detection
	checkPositions := []CollisionBox{
		// Check directly below with small height
		{
			X:      box.X,
			Y:      box.Y + 1,
			Width:  box.Width,
			Height: 3,
		},
		// Check slightly further down
		{
			X:      box.X,
			Y:      box.Y + 2,
			Width:  box.Width,
			Height: 2,
		},
		// Check if already overlapping with ground (for when falling into tiles)
		{
			X:      box.X,
			Y:      box.Y + box.Height - 5, // Check bottom 5 pixels of hitbox
			Width:  box.Width,
			Height: 5,
		},
	}

	// If any of these positions have collision, we're on ground
	for _, checkBox := range checkPositions {
		if cs.CheckCollisionAtPoint(checkBox) {
			return true
		}
	}

	return false
}

// GetSafePosition finds a safe position near the given coordinates
func (cs *CollisionSystem) GetSafePosition(box CollisionBox) (float64, float64, bool) {
	if cs.TileMap == nil {
		return box.X, box.Y, true
	}

	// If current position is safe, return it
	if !cs.CheckCollisionAtPoint(box) {
		return box.X, box.Y, true
	}

	// Try positions around the current position
	offsets := []struct{ dx, dy float64 }{
		{0, -1}, {0, -2}, {0, -3}, {0, -4}, {0, -5}, // Try moving up first
		{-1, 0}, {1, 0}, {-2, 0}, {2, 0}, // Then horizontal
		{0, 1}, {0, 2}, // Then down
		{-1, -1}, {1, -1}, {-1, 1}, {1, 1}, // Diagonal directions
	}

	for distance := 1.0; distance <= 64.0; distance += 2.0 {
		for _, offset := range offsets {
			testBox := CollisionBox{
				X:      box.X + offset.dx*distance,
				Y:      box.Y + offset.dy*distance,
				Width:  box.Width,
				Height: box.Height,
			}

			if !cs.CheckCollisionAtPoint(testBox) {
				return testBox.X, testBox.Y, true
			}
		}
	}

	// No safe position found
	return box.X, box.Y, false
}

// SlideMovement attempts to slide along walls when blocked
func (cs *CollisionSystem) SlideMovement(from, to CollisionBox) CollisionInfo {
	info := cs.CheckMovement(from, to)

	if !info.HasCollision {
		return info
	}

	// Try horizontal movement only
	horizontalOnly := CollisionBox{
		X:      to.X,
		Y:      from.Y,
		Width:  to.Width,
		Height: to.Height,
	}

	horizontalInfo := cs.CheckMovement(from, horizontalOnly)
	if !horizontalInfo.HasCollision {
		return horizontalInfo
	}

	// Try vertical movement only
	verticalOnly := CollisionBox{
		X:      from.X,
		Y:      to.Y,
		Width:  to.Width,
		Height: to.Height,
	}

	verticalInfo := cs.CheckMovement(from, verticalOnly)
	if !verticalInfo.HasCollision {
		return verticalInfo
	}

	// Both directions blocked, return original blocked info
	return info
}

// GameObject interface for objects that can collide
type GameObject interface {
	GetCollisionBox() CollisionBox
	SetPosition(x, y float64)
	OnCollision(info CollisionInfo)
}

// UpdateGameObject updates a game object with collision detection
func (cs *CollisionSystem) UpdateGameObject(obj GameObject, deltaX, deltaY float64) {
	currentBox := obj.GetCollisionBox()
	targetBox := CollisionBox{
		X:      currentBox.X + deltaX,
		Y:      currentBox.Y + deltaY,
		Width:  currentBox.Width,
		Height: currentBox.Height,
	}

	info := cs.SlideMovement(currentBox, targetBox)
	obj.SetPosition(info.NewX, info.NewY)

	if info.HasCollision {
		obj.OnCollision(info)
	}
}
