package src

import (
	"github.com/temidaradev/ebijam25/assets"
)

type CollisionSystem struct {
	TileMap *assets.TileMap
}

func NewCollisionSystem(tileMap *assets.TileMap) *CollisionSystem {
	return &CollisionSystem{
		TileMap: tileMap,
	}
}

type CollisionBox struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

type CollisionInfo struct {
	HasCollision bool
	NewX         float64
	NewY         float64
	HitWall      bool
	HitGround    bool
	HitCeiling   bool
}

func (cs *CollisionSystem) CheckMovement(from, to CollisionBox) CollisionInfo {
	info := CollisionInfo{
		HasCollision: false,
		NewX:         to.X,
		NewY:         to.Y,
	}
	if cs.TileMap == nil {
		return info
	}
	result := cs.TileMap.CheckMovementAdvanced(from.X, from.Y, to.X, to.Y, to.Width, to.Height)
	info.HasCollision = result.HasCollision
	info.NewX = result.AdjustedX
	info.NewY = result.AdjustedY
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

func (cs *CollisionSystem) CheckCollisionAtPoint(box CollisionBox) bool {
	if cs.TileMap == nil {
		return false
	}
	return cs.TileMap.CheckCollision(box.X, box.Y, box.Width, box.Height)
}

func (cs *CollisionSystem) IsOnGround(box CollisionBox) bool {
	if cs.TileMap == nil {
		return false
	}
	checkPositions := []CollisionBox{
		{X: box.X, Y: box.Y + 1, Width: box.Width, Height: 3},
		{X: box.X, Y: box.Y + 2, Width: box.Width, Height: 2},
		{X: box.X, Y: box.Y + box.Height - 5, Width: box.Width, Height: 5},
	}
	for _, checkBox := range checkPositions {
		if cs.CheckCollisionAtPoint(checkBox) {
			return true
		}
	}
	return false
}

func (cs *CollisionSystem) GetSafePosition(box CollisionBox) (float64, float64, bool) {
	if cs.TileMap == nil {
		return box.X, box.Y, true
	}
	if !cs.CheckCollisionAtPoint(box) {
		return box.X, box.Y, true
	}
	offsets := []struct{ dx, dy float64 }{
		{0, -1}, {0, -2}, {0, -3}, {0, -4}, {0, -5},
		{-1, 0}, {1, 0}, {-2, 0}, {2, 0},
		{0, 1}, {0, 2},
		{-1, -1}, {1, -1}, {-1, 1}, {1, 1},
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
	return box.X, box.Y, false
}

func (cs *CollisionSystem) SlideMovement(from, to CollisionBox) CollisionInfo {
	info := cs.CheckMovement(from, to)
	if !info.HasCollision {
		return info
	}
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
	return info
}

type GameObject interface {
	GetCollisionBox() CollisionBox
	SetPosition(x, y float64)
	OnCollision(info CollisionInfo)
}

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
