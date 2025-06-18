package assets

import (
	"log"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lafriks/go-tiled"
	"github.com/solarlune/resolv"
	"github.com/temidaradev/esset/v2"
)

type TileMap struct {
	Image          *ebiten.Image
	Map            *tiled.Map
	TileWidth      int
	TileHeight     int
	MapWidth       int
	MapHeight      int
	PixelWidth     int
	PixelHeight    int
	CollisionSpace *resolv.Space
}

var (
	DesertTileMap *TileMap
	ForestTileMap *TileMap
	CaveTileMap   *TileMap
)

func InitTileMaps() {
	DesertTileMap = LoadTileMap("images/backgrounds/desert-tiles/desert.tmx")
	// Add other tilemaps as needed
	// ForestTileMap = LoadTileMap("images/backgrounds/forest-tiles/forest.tmx")
	// CaveTileMap = LoadTileMap("images/backgrounds/cave-tiles/cave.tmx")
}

func LoadTileMap(mapPath string) *TileMap {
	gameMap, err := tiled.LoadFile(mapPath, tiled.WithFileSystem(assets))
	if err != nil {
		log.Printf("Failed to parse map %s: %v", mapPath, err)
		return nil
	}

	tileMap := &TileMap{
		Map:            gameMap,
		TileWidth:      gameMap.TileWidth,
		TileHeight:     gameMap.TileHeight,
		MapWidth:       gameMap.Width,
		MapHeight:      gameMap.Height,
		PixelWidth:     gameMap.Width * gameMap.TileWidth,
		PixelHeight:    gameMap.Height * gameMap.TileHeight,
		CollisionSpace: resolv.NewSpace(gameMap.Width*gameMap.TileWidth, gameMap.Height*gameMap.TileHeight, 32, 32),
	}

	tileMap.Image = renderTileMapToImage(gameMap)
	tileMap.createCollisionObjects()

	return tileMap
}

func renderTileMapToImage(gameMap *tiled.Map) *ebiten.Image {
	mapImage := ebiten.NewImage(gameMap.Width*gameMap.TileWidth, gameMap.Height*gameMap.TileHeight)

	tileImages := make(map[uint32]*ebiten.Image)

	for _, tileset := range gameMap.Tilesets {
		log.Printf("Processing tileset: %s, FirstGID: %d, TileCount: %d", tileset.Name, tileset.FirstGID, tileset.TileCount)
		if tileset.Tiles != nil {
			for _, tile := range tileset.Tiles {
				if tile.Image != nil {
					// Construct the tile image path
					tilePath := filepath.Join(filepath.Dir("images/backgrounds/desert-tiles/desert.tmx"), tile.Image.Source)
					tileImg := esset.GetAsset(assets, tilePath)
					if tileImg != nil {
						globalID := tileset.FirstGID + tile.ID
						tileImages[globalID] = tileImg
						log.Printf("Loaded tile image: %s for local tile ID %d -> global ID %d", tile.Image.Source, tile.ID, globalID)
					} else {
						log.Printf("Failed to load tile image: %s", tilePath)
					}
				}
			}
		}
	}

	// Render each layer
	for _, layer := range gameMap.Layers {
		log.Printf("Processing layer: %s", layer.Name)
		if len(layer.Tiles) > 0 && layer.Visible {
			log.Printf("Rendering tile layer: %s with %d tiles", layer.Name, len(layer.Tiles))
			renderTileLayerFromImages(mapImage, layer, gameMap, tileImages)
		} else {
			log.Printf("Skipping layer %s: has %d tiles, visible: %v", layer.Name, len(layer.Tiles), layer.Visible)
		}
	}

	return mapImage
}

func renderTileLayerFromImages(mapImage *ebiten.Image, layer *tiled.Layer, gameMap *tiled.Map, tileImages map[uint32]*ebiten.Image) {
	for y := 0; y < gameMap.Height; y++ {
		for x := 0; x < gameMap.Width; x++ {
			tileIndex := y*gameMap.Width + x
			if tileIndex >= len(layer.Tiles) {
				continue
			}

			tile := layer.Tiles[tileIndex]
			if tile.ID == 0 {
				continue
			}

			// DEBUG: Log first few tiles to see what's happening
			if tileIndex < 10 {
				log.Printf("DEBUG: Tile at index %d (pos %d,%d): ID=%d", tileIndex, x, y, tile.ID)
			}

			// Add +1 to the tile ID to get the correct tile
			correctedTileID := tile.ID + 1
			tileImage := tileImages[correctedTileID]
			if tileImage == nil {
				if tileIndex < 10 { // Only log first few missing tiles to avoid spam
					log.Printf("Warning: No image found for tile ID %d (corrected to %d) at position (%d, %d)", tile.ID, correctedTileID, x, y)
				}
				continue
			}

			// Calculate destination position
			dstX := x * gameMap.TileWidth
			dstY := y * gameMap.TileHeight

			// Draw the tile
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(dstX), float64(dstY))

			mapImage.DrawImage(tileImage, op)
		}
	}
}

// Draw renders the tilemap to the screen with camera transformation
func (tm *TileMap) Draw(screen *ebiten.Image, cameraX, cameraY, screenWidth, screenHeight float64) {
	if tm.Image == nil {
		return
	}

	// Calculate which portion of the tilemap is visible
	// Apply camera offset
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-cameraX, -cameraY)

	// For now, draw the entire tilemap
	// In a more optimized version, you would only draw the visible portion
	screen.DrawImage(tm.Image, op)
}

// GetBounds returns the bounds of the tilemap in world coordinates
func (tm *TileMap) GetBounds() (x, y, width, height float64) {
	return 0, 0, float64(tm.PixelWidth), float64(tm.PixelHeight)
}

// GetTileAt returns the tile ID at the given world coordinates
func (tm *TileMap) GetTileAt(worldX, worldY float64) uint32 {
	if tm.Map == nil {
		return 0
	}

	tileX := int(worldX) / tm.TileWidth
	tileY := int(worldY) / tm.TileHeight

	if tileX < 0 || tileX >= tm.MapWidth || tileY < 0 || tileY >= tm.MapHeight {
		return 0
	}

	// Get the first tile layer
	for _, layer := range tm.Map.Layers {
		if len(layer.Tiles) > 0 {
			tileIndex := tileY*tm.MapWidth + tileX
			if tileIndex < len(layer.Tiles) {
				return layer.Tiles[tileIndex].ID
			}
		}
	}

	return 0
}

// createCollisionObjects creates collision boxes for solid tiles
func (tm *TileMap) createCollisionObjects() {
	if tm.Map == nil || tm.CollisionSpace == nil {
		return
	}

	collisionCount := 0
	totalTiles := 0

	// Process each layer to create collision objects
	for _, layer := range tm.Map.Layers {
		if len(layer.Tiles) > 0 && layer.Visible {
			log.Printf("Processing layer: %s with %d tiles", layer.Name, len(layer.Tiles))
			for y := 0; y < tm.MapHeight; y++ {
				for x := 0; x < tm.MapWidth; x++ {
					tileIndex := y*tm.MapWidth + x
					if tileIndex >= len(layer.Tiles) {
						continue
					}

					tile := layer.Tiles[tileIndex]
					totalTiles++
					if tile.ID == 0 {
						continue // Empty tile
					}

					// Check if this tile should have collision using helper function
					if IsTileSolid(tile.ID) {
						// Create a collision rectangle for this tile
						rect := resolv.NewRectangle(
							float64(x*tm.TileWidth),
							float64(y*tm.TileHeight),
							float64(tm.TileWidth),
							float64(tm.TileHeight))
						tm.CollisionSpace.Add(rect)
						collisionCount++
					}
				}
			}
		}
	}

	log.Printf("Created collision objects for tilemap: %d collision objects from %d total tiles", collisionCount, totalTiles)
}

// RecreateCollisionObjects clears and recreates all collision objects
func (tm *TileMap) RecreateCollisionObjects() {
	if tm.Map == nil {
		return
	}

	// Recreate the collision space to clear all existing objects
	tm.CollisionSpace = resolv.NewSpace(tm.Map.Width*tm.Map.TileWidth, tm.Map.Height*tm.Map.TileHeight, 32, 32)

	// Recreate them with current rules
	tm.createCollisionObjects()
}

// CheckCollision checks for collision with a rectangle at the given position
func (tm *TileMap) CheckCollision(x, y, width, height float64) bool {
	if tm.CollisionSpace == nil {
		return false
	}

	// Create a temporary rectangle for collision checking
	checkRect := resolv.NewRectangle(x, y, width, height)

	// Check for intersection with any shapes in the collision space
	for _, shape := range tm.CollisionSpace.Shapes() {
		if intersections := checkRect.Intersection(shape); len(intersections.Intersections) > 0 {
			return true
		}
	}
	return false
}

// CheckMovement checks if movement to a new position is valid and returns adjusted position
func (tm *TileMap) CheckMovement(fromX, fromY, toX, toY, width, height float64) (float64, float64, bool) {
	if tm.CollisionSpace == nil {
		return toX, toY, true
	}

	// Create a rectangle at the target position
	targetRect := resolv.NewRectangle(toX, toY, width, height)

	// Check for intersections
	for _, shape := range tm.CollisionSpace.Shapes() {
		if intersections := targetRect.Intersection(shape); len(intersections.Intersections) > 0 {
			// There's a collision, prevent the movement
			return fromX, fromY, false
		}
	}

	// No collision, movement is valid
	return toX, toY, true
}

// CollisionResult represents the result of a collision check
type CollisionResult struct {
	HasCollision bool
	AdjustedX    float64
	AdjustedY    float64
	CollisionX   bool
	CollisionY   bool
	NormalX      float64
	NormalY      float64
}

// CheckMovementAdvanced performs advanced collision detection with proper resolution
func (tm *TileMap) CheckMovementAdvanced(fromX, fromY, toX, toY, width, height float64) CollisionResult {
	result := CollisionResult{
		HasCollision: false,
		AdjustedX:    toX,
		AdjustedY:    toY,
		CollisionX:   false,
		CollisionY:   false,
	}

	if tm.CollisionSpace == nil {
		return result
	}

	// Check horizontal movement first
	horizontalRect := resolv.NewRectangle(toX, fromY, width, height)
	hasHorizontalCollision := false

	for _, shape := range tm.CollisionSpace.Shapes() {
		if intersections := horizontalRect.Intersection(shape); len(intersections.Intersections) > 0 {
			hasHorizontalCollision = true
			result.HasCollision = true
			result.CollisionX = true
			break
		}
	}

	// Check vertical movement
	verticalRect := resolv.NewRectangle(fromX, toY, width, height)
	hasVerticalCollision := false

	for _, shape := range tm.CollisionSpace.Shapes() {
		if intersections := verticalRect.Intersection(shape); len(intersections.Intersections) > 0 {
			hasVerticalCollision = true
			result.HasCollision = true
			result.CollisionY = true
			break
		}
	}

	// Resolve collisions by adjusting position
	if hasHorizontalCollision {
		result.AdjustedX = fromX // Block horizontal movement completely
	} else {
		result.AdjustedX = toX // Allow horizontal movement
	}

	if hasVerticalCollision {
		// For vertical collision, we need to position the player properly
		if toY > fromY {
			// Player is falling down - find the top-most tile they would hit
			bestY := toY
			targetRect := resolv.NewRectangle(fromX, toY, width, height)

			for _, shape := range tm.CollisionSpace.Shapes() {
				if intersections := targetRect.Intersection(shape); len(intersections.Intersections) > 0 {
					if rect, ok := shape.(*resolv.ConvexPolygon); ok {
						shapePos := rect.Position()
						// Position player on top of this tile
						tileTop := shapePos.Y - height
						if tileTop < bestY {
							bestY = tileTop
						}
					}
				}
			}
			result.AdjustedY = bestY
		} else {
			// Player is moving up, block movement
			result.AdjustedY = fromY
		}
	} else {
		result.AdjustedY = toY // Allow vertical movement
	}

	return result
}

// CheckGroundCollision specifically checks for ground collision (useful for gravity)
func (tm *TileMap) CheckGroundCollision(x, y, width, height float64) (bool, float64) {
	if tm.CollisionSpace == nil {
		return false, y
	}

	// Check slightly below the current position
	checkRect := resolv.NewRectangle(x, y+1, width, height)

	for _, shape := range tm.CollisionSpace.Shapes() {
		if intersections := checkRect.Intersection(shape); len(intersections.Intersections) > 0 {
			// Found ground collision, return the top of the collision shape
			if rect, ok := shape.(*resolv.ConvexPolygon); ok {
				shapePos := rect.Position()
				return true, shapePos.Y - height
			}
		}
	}

	return false, y
}

// GetCollisionShapes returns all collision shapes for debugging
func (tm *TileMap) GetCollisionShapes() []*resolv.ConvexPolygon {
	if tm.CollisionSpace == nil {
		return nil
	}

	var shapes []*resolv.ConvexPolygon
	for _, shape := range tm.CollisionSpace.Shapes() {
		if rect, ok := shape.(*resolv.ConvexPolygon); ok {
			shapes = append(shapes, rect)
		}
	}
	return shapes
}

// Helper function to check if a specific tile ID should be solid
func IsTileSolid(tileID uint32) bool {
	// Every non-empty tile is now treated as ground/solid
	// This removes borders and makes all tiles act as collision surfaces
	return tileID > 0 // Any tile with ID > 0 is solid
}

// GetTileCollisionInfo returns collision information for debugging
func (tm *TileMap) GetTileCollisionInfo(worldX, worldY float64) (uint32, bool) {
	tileID := tm.GetTileAt(worldX, worldY)
	return tileID, IsTileSolid(tileID)
}

// Legacy function for backward compatibility
func GetMap() *TileMap {
	return DesertTileMap
}
