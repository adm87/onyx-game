package tiled

import (
	"github.com/adm87/onyx/pkg/engine/geom"
)

type Tile struct {
	id      uint32
	tileset int
}

func (t Tile) ID() uint32 {
	return uint32(t.id & GidMask)
}

func (t Tile) Tileset() int {
	return t.tileset
}

func (t Tile) Flags() uint32 {
	return uint32(t.id &^ GidMask)
}

func (t Tile) FlippedHorizontally() bool {
	return t.Flags()&FlippedHorizontallyFlag != 0
}

func (t Tile) FlippedVertically() bool {
	return t.Flags()&FlippedVerticallyFlag != 0
}

func (t Tile) FlippedDiagonally() bool {
	return t.Flags()&FlippedDiagonallyFlag != 0
}

func (t Tile) RotatedHexagonal120() bool {
	return t.Flags()&RotatedHexagonal120Flag != 0
}

type Tilemap struct {
	layers     int
	bounds     geom.AABB
	tileBounds geom.AABB
	tiles      []Tile // Flattened array of tiles, ordered by layer and then by position
}

func (t *Tilemap) GetTileIndex(layer, x, y int) int {
	width := int(t.tileBounds.Width())
	height := int(t.tileBounds.Height())
	return (layer * width * height) + (y * width) + x
}

func (t *Tilemap) GetTile(layer, x, y int) (Tile, int, bool) {
	wx := x - int(t.tileBounds.Min.X)
	wy := y - int(t.tileBounds.Min.Y)
	w, h := int(t.tileBounds.Width()), int(t.tileBounds.Height())
	if layer < 0 || layer >= t.layers || wx < 0 || wx >= w || wy < 0 || wy >= h {
		return Tile{0, 0}, 0, false
	}
	index := t.GetTileIndex(layer, wx, wy)
	return t.tiles[index], index, true
}

func (t *Tilemap) Bounds() geom.AABB {
	return t.bounds
}

func (t *Tilemap) TileBounds() geom.AABB {
	return t.tileBounds
}

func (t *Tilemap) Layers() int {
	return t.layers
}

func buildTilemap(tmx *Tmx) (*Tilemap, error) {
	min, max := findMapSize(tmx)
	tileBounds := geom.AABB{Min: min, Max: max}
	bounds := geom.AABB{
		Min: geom.Vec2{
			X: tileBounds.Min.X * float64(tmx.TileWidth),
			Y: tileBounds.Min.Y * float64(tmx.TileHeight),
		},
		Max: geom.Vec2{
			X: tileBounds.Max.X * float64(tmx.TileWidth),
			Y: tileBounds.Max.Y * float64(tmx.TileHeight),
		},
	}
	size := int(tileBounds.Width() * tileBounds.Height())
	tilemap := &Tilemap{
		tileBounds: tileBounds,
		bounds:     bounds,
		layers:     len(tmx.Layers),
		tiles:      make([]Tile, size*len(tmx.Layers)),
	}
	for i, layer := range tmx.Layers {
		if err := buildTilemapLayer(layer, tilemap, tmx.Tilesets, i, size, tileBounds); err != nil {
			return nil, err
		}
	}
	return tilemap, nil
}

func buildTilemapLayer(layer TmxLayer, tilemap *Tilemap, tilesets []TmxTileset, i, size int, bounds geom.AABB) error {
	var tiles []Tile
	var err error

	if len(layer.Data.Chunks) > 0 {
		mapWidth := int(bounds.Width())
		layerOffset := i * size
		for _, chunk := range layer.Data.Chunks {
			tiles, err = decodeContent(layer.Data.Encoding, layer.Data.Compression, chunk.Content)
			if err != nil {
				return err
			}
			for i := range tiles {
				_, j := NearestTileset(tilesets, tiles[i].ID())
				tiles[i].tileset = j
			}
			chunkX := chunk.X - int(bounds.Min.X)
			chunkY := chunk.Y - int(bounds.Min.Y)
			chunkOffset := layerOffset + (chunkY * mapWidth) + chunkX
			for row := 0; row < chunk.Height; row++ {
				src := row * chunk.Width
				dst := chunkOffset + row*mapWidth
				copy(tilemap.tiles[dst:dst+chunk.Width], tiles[src:src+chunk.Width])
			}
		}
	} else {
		tiles, err = decodeContent(layer.Data.Encoding, layer.Data.Compression, layer.Data.Content)
		if err != nil {
			return err
		}
		for i := range tiles {
			_, j := NearestTileset(tilesets, tiles[i].ID())
			tiles[i].tileset = j
		}
		copy(tilemap.tiles[i*size:(i+1)*size], tiles)
	}
	return nil
}

func findMapSize(tmx *Tmx) (geom.Vec2, geom.Vec2) {
	if len(tmx.Layers) == 0 {
		return geom.Vec2{}, geom.Vec2{}
	}

	minX, minY, maxX, maxY := findLayerSize(tmx.Layers[0])
	for _, layer := range tmx.Layers[1:] {
		layerMinX, layerMinY, layerMaxX, layerMaxY := findLayerSize(layer)
		if layerMinX < minX {
			minX = layerMinX
		}
		if layerMinY < minY {
			minY = layerMinY
		}
		if layerMaxX > maxX {
			maxX = layerMaxX
		}
		if layerMaxY > maxY {
			maxY = layerMaxY
		}
	}

	return geom.Vec2{
			X: float64(minX),
			Y: float64(minY),
		}, geom.Vec2{
			X: float64(maxX),
			Y: float64(maxY),
		}
}

func findLayerSize(layer TmxLayer) (minX, minY, maxX, maxY int) {
	if len(layer.Data.Chunks) == 0 {
		return 0, 0, layer.Width, layer.Height
	}

	first := layer.Data.Chunks[0]
	minX, minY = first.X, first.Y
	maxX, maxY = first.X+first.Width, first.Y+first.Height

	for _, chunk := range layer.Data.Chunks[1:] {
		if chunk.X < minX {
			minX = chunk.X
		}
		if chunk.Y < minY {
			minY = chunk.Y
		}
		if chunk.X+chunk.Width > maxX {
			maxX = chunk.X + chunk.Width
		}
		if chunk.Y+chunk.Height > maxY {
			maxY = chunk.Y + chunk.Height
		}
	}

	return minX, minY, maxX, maxY
}
