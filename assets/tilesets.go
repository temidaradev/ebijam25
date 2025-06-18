package assets

import (
	"log"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lafriks/go-tiled"
	"github.com/temidaradev/esset/v2"
)

// TileMap represents a rendered tilemap that can be drawn to the screen
type TileMap struct {
	Image       *ebiten.Image
	Map         *tiled.Map
	TileWidth   int
	TileHeight  int
	MapWidth    int
	MapHeight   int
	PixelWidth  int
	PixelHeight int
}

// Global tilemap variables
var (
	DesertTileMap *TileMap
	ForestTileMap *TileMap
	CaveTileMap   *TileMap
)

// InitTileMaps initializes all tilemaps
func InitTileMaps() {
	DesertTileMap = LoadTileMap("images/backgrounds/desert-tiles/desert.tmx")
	// Add other tilemaps as needed
	// ForestTileMap = LoadTileMap("images/backgrounds/forest-tiles/forest.tmx")
	// CaveTileMap = LoadTileMap("images/backgrounds/cave-tiles/cave.tmx")
}

// LoadTileMap loads a Tiled map file and renders it to an image
func LoadTileMap(mapPath string) *TileMap {
	// Load the map from embedded assets using WithFileSystem option
	gameMap, err := tiled.LoadFile(mapPath, tiled.WithFileSystem(assets))
	if err != nil {
		log.Printf("Failed to parse map %s: %v", mapPath, err)
		return nil
	}

	// Create the tilemap structure
	tileMap := &TileMap{
		Map:         gameMap,
		TileWidth:   gameMap.TileWidth,
		TileHeight:  gameMap.TileHeight,
		MapWidth:    gameMap.Width,
		MapHeight:   gameMap.Height,
		PixelWidth:  gameMap.Width * gameMap.TileWidth,
		PixelHeight: gameMap.Height * gameMap.TileHeight,
	}

	// Pre-render the tilemap to an image for better performance
	tileMap.Image = renderTileMapToImage(gameMap)

	return tileMap
}

// renderTileMapToImage renders the entire tilemap to a single image
func renderTileMapToImage(gameMap *tiled.Map) *ebiten.Image {
	// Create an image to render the tilemap onto
	mapImage := ebiten.NewImage(gameMap.Width*gameMap.TileWidth, gameMap.Height*gameMap.TileHeight)

	// Load individual tile images for collection tilesets
	tileImages := make(map[uint32]*ebiten.Image)

	for _, tileset := range gameMap.Tilesets {
		log.Printf("Processing tileset: %s, FirstGID: %d, TileCount: %d", tileset.Name, tileset.FirstGID, tileset.TileCount)

		// Handle collection of images tileset (individual tile images)
		if tileset.Tiles != nil {
			for _, tile := range tileset.Tiles {
				if tile.Image != nil {
					// Construct the tile image path
					tilePath := filepath.Join(filepath.Dir("images/backgrounds/desert-tiles/desert.tmx"), tile.Image.Source)
					tileImg := esset.GetAsset(assets, tilePath)
					if tileImg != nil {
						// The global ID is the tileset's FirstGID + the tile's local ID
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

		// Check if this is a tile layer by checking if it has tiles
		if len(layer.Tiles) > 0 && layer.Visible {
			log.Printf("Rendering tile layer: %s with %d tiles", layer.Name, len(layer.Tiles))
			renderTileLayerFromImages(mapImage, layer, gameMap, tileImages)
		} else {
			log.Printf("Skipping layer %s: has %d tiles, visible: %v", layer.Name, len(layer.Tiles), layer.Visible)
		}
	}

	return mapImage
}

// renderTileLayerFromImages renders a single tile layer using individual tile images
func renderTileLayerFromImages(mapImage *ebiten.Image, layer *tiled.Layer, gameMap *tiled.Map, tileImages map[uint32]*ebiten.Image) {
	for y := 0; y < gameMap.Height; y++ {
		for x := 0; x < gameMap.Width; x++ {
			tileIndex := y*gameMap.Width + x
			if tileIndex >= len(layer.Tiles) {
				continue
			}

			tile := layer.Tiles[tileIndex]
			if tile.ID == 0 {
				continue // Empty tile
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

// Legacy function for backward compatibility
func GetMap() *TileMap {
	return DesertTileMap
}
