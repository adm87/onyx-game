package tiled

import (
	"math"

	"github.com/adm87/onyx/pkg/engine"
	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/adm87/onyx/pkg/modules/ecs"
	"github.com/adm87/onyx/pkg/modules/ecs/renderer"
	"github.com/adm87/onyx/pkg/modules/images"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

type TiledECSRenderer struct {
	imageAssets *images.ImageAssets
	tiledAssets *TiledAssets

	screen engine.Screen

	lastViewport geom.AABB

	tasks    []*engine.RenderingTask
	drawOpts *ebiten.DrawImageOptions
}

func NewTiledECSRenderer(assets *TiledAssets) *TiledECSRenderer {
	return &TiledECSRenderer{
		tiledAssets: assets,
		tasks:       make([]*engine.RenderingTask, 0, 10),
		drawOpts:    &ebiten.DrawImageOptions{},
	}
}

func (r *TiledECSRenderer) GetRenderingTasks(
	entry *donburi.Entry,
	renderer *renderer.RendererModel,
	viewport geom.AABB,
	viewMatrix ebiten.GeoM,
	pool *ecs.RenderingPool,
	renderingTasks []*engine.RenderingTask) []*engine.RenderingTask {
	r.tasks = r.tasks[:0]

	tilemapHandle := GetTilemapHandle(entry)

	tilemap, exists := r.tiledAssets.GetTilemap(tilemapHandle)
	if !exists {
		return renderingTasks
	}

	tmx, exists := r.tiledAssets.GetTmx(tilemapHandle)
	if !exists {
		return renderingTasks
	}

	minTileX := int(math.Floor(viewport.Min.X / float64(tmx.TileWidth)))
	maxTileX := int(math.Floor(viewport.Max.X / float64(tmx.TileWidth)))
	minTileY := int(math.Floor(viewport.Min.Y / float64(tmx.TileHeight)))
	maxTileY := int(math.Floor(viewport.Max.Y / float64(tmx.TileHeight)))

	viewChanged := !r.lastViewport.Equals(viewport)
	r.lastViewport = viewport

	screenSize := r.screen.Size()
	for i := range tilemap.Layers() {
		if !tmx.Layers[i].Visible {
			continue
		}

		buffer, resized := r.tiledAssets.GetTilemapBuffer(tilemapHandle, int(screenSize.X), int(screenSize.Y), i)
		if resized || viewChanged {
			buffer.Clear()

			r.drawTilemapLayer(
				buffer,
				tilemap, i,
				tmx.TileWidth, tmx.TileHeight,
				minTileX, maxTileX,
				minTileY, maxTileY,
				tmx.Tilesets,
				viewMatrix,
			)
		}

		task := pool.Get()
		task.Buffer = buffer
		task.Layer = renderer.Layer
		task.ZIndex = renderer.ZIndex + float32(i)
		renderingTasks = append(renderingTasks, task)
	}

	return renderingTasks
}

func (a *TiledECSRenderer) drawTilemapLayer(
	target *ebiten.Image,
	tilemam *Tilemap,
	layerIndex int,
	cellWidth, cellHeight int,
	minTileX, maxTileX int,
	minTileY, maxTileY int,
	tilesets []TmxTileset,
	viewMatrix ebiten.GeoM) {

	for y := minTileY; y <= maxTileY; y++ {
		for x := minTileX; x <= maxTileX; x++ {
			tile, _, exists := tilemam.GetTile(layerIndex, x, y)
			if !exists || tile.ID() == 0 {
				continue
			}
			j := tile.Tileset()

			// TODO - While just a slotmap lookup, we should consider caching this somewhere.
			tsx, exists := a.tiledAssets.GetTsx(tilesets[j].Handle)
			if !exists {
				continue
			}

			a.drawOpts.GeoM.Reset()

			// TODO - Consider precomputing these transform offsets.

			// // Ref: https://doc.mapeditor.org/en/stable/reference/global-tile-ids/#tile-flipping
			if tile.FlippedDiagonally() {
				a.drawOpts.GeoM.Rotate(math.Pi * 0.5)
				a.drawOpts.GeoM.Scale(-1, 1)
				a.drawOpts.GeoM.Translate(float64(tsx.TileHeight-tsx.TileWidth), 0)
			}
			if tile.FlippedHorizontally() {
				a.drawOpts.GeoM.Scale(-1, 1)
				a.drawOpts.GeoM.Translate(float64(tsx.TileWidth), 0)
			}
			if tile.FlippedVertically() {
				a.drawOpts.GeoM.Scale(1, -1)
				a.drawOpts.GeoM.Translate(0, float64(tsx.TileHeight))
			}

			tileX, tileY := x*cellWidth, y*cellHeight
			a.drawOpts.GeoM.Translate(0, float64(cellHeight-tsx.TileHeight))
			a.drawOpts.GeoM.Translate(float64(tsx.TileOffset.X), float64(tsx.TileOffset.Y))
			a.drawOpts.GeoM.Translate(float64(tileX), float64(tileY))
			a.drawOpts.GeoM.Concat(viewMatrix)

			tileID := tile.ID() - uint32(tilesets[j].FirstGID)

			frame, exists := a.imageAssets.GetFrame(tsx.Image.Handle, int(tileID))
			if !exists {
				continue
			}

			target.DrawImage(frame, a.drawOpts)
		}
	}
}
