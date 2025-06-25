package assets

import (
	"image"
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
)

func InitTileMaps() {
	DesertTileMap = LoadTileMap("images/backgrounds/desert-tiles/desert.tmx")
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
		CollisionSpace: resolv.NewSpace(gameMap.Width*gameMap.TileWidth, gameMap.Height*gameMap.TileHeight, gameMap.TileWidth, gameMap.TileHeight),
	}

	tileMap.Image = renderTileMapToImage(gameMap, mapPath)
	tileMap.createCollisionObjects()
	return tileMap
}

func renderTileMapToImage(gameMap *tiled.Map, mapPath string) *ebiten.Image {
	mapImage := ebiten.NewImage(gameMap.Width*gameMap.TileWidth, gameMap.Height*gameMap.TileHeight)
	tileImages := make(map[uint32]*ebiten.Image)

	for _, tileset := range gameMap.Tilesets {
		if tileset.Tiles != nil {
			for _, tile := range tileset.Tiles {
				if tile.Image != nil {
					tilePath := filepath.Join(filepath.Dir(mapPath), tile.Image.Source)
					tileImg := esset.GetAsset(assets, tilePath)
					if tileImg != nil {
						globalID := tileset.FirstGID + tile.ID
						tileImages[globalID] = tileImg
					}
				}
			}
		}
		if tileset.Image != nil {
			tilesetPath := filepath.Join(filepath.Dir(mapPath), tileset.Image.Source)
			tilesetImg := esset.GetAsset(assets, tilesetPath)
			if tilesetImg != nil {
				tileWidth := tileset.TileWidth
				tileHeight := tileset.TileHeight
				columns := uint32(tileset.Columns)
				margin := tileset.Margin
				spacing := tileset.Spacing
				for tileID := uint32(0); tileID < uint32(tileset.TileCount); tileID++ {
					col := int(tileID % columns)
					row := int(tileID / columns)
					srcX := margin + col*(tileWidth+spacing)
					srcY := margin + row*(tileHeight+spacing)
					tileImg := tilesetImg.SubImage(image.Rect(srcX, srcY, srcX+tileWidth, srcY+tileHeight)).(*ebiten.Image)
					globalID := tileset.FirstGID + tileID
					tileImages[globalID] = tileImg
				}
			}
		}
	}
	for _, layer := range gameMap.Layers {
		if len(layer.Tiles) > 0 && layer.Visible {
			renderTileLayerFromImages(mapImage, layer, gameMap, tileImages)
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
			correctedTileID := tile.ID + 1
			tileImage := tileImages[correctedTileID]
			if tileImage == nil {
				continue
			}
			dstX := x * gameMap.TileWidth
			dstY := y * gameMap.TileHeight
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(dstX), float64(dstY))
			mapImage.DrawImage(tileImage, op)
		}
	}
}

func (tm *TileMap) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	if tm.Image == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-cameraX, -cameraY)
	screen.DrawImage(tm.Image, op)
}

func (tm *TileMap) drawTiled(screen *ebiten.Image, cameraX, cameraY, screenWidth, screenHeight float64) {
	if tm.Image == nil {
		return
	}
	tileMapWidth := float64(tm.PixelWidth)
	tileMapHeight := float64(tm.PixelHeight)
	tilesX := int((screenWidth+tileMapWidth)/tileMapWidth) + 2
	tilesY := int((screenHeight+tileMapHeight)/tileMapHeight) + 2
	startX := int(cameraX / tileMapWidth)
	startY := int(cameraY / tileMapHeight)
	for y := startY - 1; y < startY+tilesY; y++ {
		for x := startX - 1; x < startX+tilesX; x++ {
			op := &ebiten.DrawImageOptions{}
			offsetX := float64(x)*tileMapWidth - cameraX
			offsetY := float64(y)*tileMapHeight - cameraY
			op.GeoM.Translate(offsetX, offsetY)
			screen.DrawImage(tm.Image, op)
		}
	}
}

func (tm *TileMap) GetBounds() (x, y, width, height float64) {
	return 0, 0, float64(tm.PixelWidth), float64(tm.PixelHeight)
}

func (tm *TileMap) GetTileAt(worldX, worldY float64) uint32 {
	if tm.Map == nil {
		return 0
	}
	tileX := int(worldX) / tm.TileWidth
	tileY := int(worldY) / tm.TileHeight
	if tileX < 0 || tileX >= tm.MapWidth || tileY < 0 || tileY >= tm.MapHeight {
		return 0
	}
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

func (tm *TileMap) createCollisionObjects() {
	if tm.Map == nil || tm.CollisionSpace == nil {
		return
	}
	collisionCount := 0
	totalTiles := 0
	for _, layer := range tm.Map.Layers {
		if len(layer.Tiles) > 0 && layer.Visible {
			for y := 0; y < tm.MapHeight; y++ {
				for x := 0; x < tm.MapWidth; x++ {
					tileIndex := y*tm.MapWidth + x
					if tileIndex >= len(layer.Tiles) {
						continue
					}
					tile := layer.Tiles[tileIndex]
					totalTiles++
					if tile.ID == 0 {
						continue
					}
					if IsTileSolid(tile.ID) {
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

func (tm *TileMap) CheckCollision(x, y, width, height float64) bool {
	if tm.CollisionSpace == nil {
		return false
	}
	checkRect := resolv.NewRectangle(x, y, width, height)
	for _, shape := range tm.CollisionSpace.Shapes() {
		if intersections := checkRect.Intersection(shape); len(intersections.Intersections) > 0 {
			return true
		}
	}
	return false
}

func (tm *TileMap) checkTiledCollision(x, y, width, height float64) bool {
	tileMapWidth := float64(tm.PixelWidth)
	tileMapHeight := float64(tm.PixelHeight)
	leftTileX := int((x) / tileMapWidth)
	rightTileX := int((x + width) / tileMapWidth)
	topTileY := int((y) / tileMapHeight)
	bottomTileY := int((y + height) / tileMapHeight)
	for tileY := topTileY; tileY <= bottomTileY; tileY++ {
		for tileX := leftTileX; tileX <= rightTileX; tileX++ {
			localX := x - float64(tileX)*tileMapWidth
			localY := y - float64(tileY)*tileMapHeight
			checkRect := resolv.NewRectangle(localX, localY, width, height)
			for _, shape := range tm.CollisionSpace.Shapes() {
				if intersections := checkRect.Intersection(shape); len(intersections.Intersections) > 0 {
					return true
				}
			}
		}
	}
	return false
}

func (tm *TileMap) CheckMovement(fromX, fromY, toX, toY, width, height float64) (float64, float64, bool) {
	if tm.CollisionSpace == nil {
		return toX, toY, true
	}
	targetRect := resolv.NewRectangle(toX, toY, width, height)
	for _, shape := range tm.CollisionSpace.Shapes() {
		if intersections := targetRect.Intersection(shape); len(intersections.Intersections) > 0 {
			return fromX, fromY, false
		}
	}
	return toX, toY, true
}

type CollisionResult struct {
	HasCollision bool
	AdjustedX    float64
	AdjustedY    float64
	CollisionX   bool
	CollisionY   bool
	NormalX      float64
	NormalY      float64
}

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
	if hasHorizontalCollision {
		result.AdjustedX = fromX
	} else {
		result.AdjustedX = toX
	}
	if hasVerticalCollision {
		if toY > fromY {
			bestY := toY
			targetRect := resolv.NewRectangle(fromX, toY, width, height)
			for _, shape := range tm.CollisionSpace.Shapes() {
				if intersections := targetRect.Intersection(shape); len(intersections.Intersections) > 0 {
					if rect, ok := shape.(*resolv.ConvexPolygon); ok {
						shapePos := rect.Position()
						tileTop := shapePos.Y - height
						if tileTop < bestY {
							bestY = tileTop
						}
					}
				}
			}
			result.AdjustedY = bestY
		} else {
			result.AdjustedY = fromY
		}
	} else {
		result.AdjustedY = toY
	}
	return result
}

func (tm *TileMap) CheckGroundCollision(x, y, width, height float64) (bool, float64) {
	if tm.CollisionSpace == nil {
		return false, y
	}
	checkRect := resolv.NewRectangle(x, y+1, width, height)
	for _, shape := range tm.CollisionSpace.Shapes() {
		if intersections := checkRect.Intersection(shape); len(intersections.Intersections) > 0 {
			if rect, ok := shape.(*resolv.ConvexPolygon); ok {
				shapePos := rect.Position()
				return true, shapePos.Y - height
			}
		}
	}
	return false, y
}

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

func IsTileSolid(tileID uint32) bool {
	return tileID > 0
}

func (tm *TileMap) GetTileCollisionInfo(worldX, worldY float64) (uint32, bool) {
	tileID := tm.GetTileAt(worldX, worldY)
	return tileID, IsTileSolid(tileID)
}
