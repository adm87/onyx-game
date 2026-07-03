package tiled

import (
	"encoding/xml"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/adm87/onyx/pkg/engine"
	"github.com/adm87/onyx/pkg/engine/assert"
	"github.com/adm87/onyx/pkg/engine/file"
	"github.com/adm87/onyx/pkg/modules/images"
	"github.com/hajimehoshi/ebiten/v2"
)

type TiledAssets struct {
	supportedExtensions []file.FileExt

	tmxStore       file.FileStore[*Tmx]
	tsxStore       file.FileStore[*Tsx]
	tilemaps       map[uint64]*Tilemap
	tilemapBuffers map[uint64][]*ebiten.Image

	imageAssets *images.ImageAssets
}

func NewTiledAssets() *TiledAssets {
	return &TiledAssets{
		supportedExtensions: []file.FileExt{".tmx", ".tsx"},
		tmxStore:            file.NewFileStore[*Tmx](0),
		tsxStore:            file.NewFileStore[*Tsx](0),
		tilemaps:            make(map[uint64]*Tilemap),
		tilemapBuffers:      make(map[uint64][]*ebiten.Image),
	}
}

func (a *TiledAssets) GetTmx(handle uint64) (*Tmx, bool) {
	return a.tmxStore.Get(handle)
}

func (a *TiledAssets) GetTsx(handle uint64) (*Tsx, bool) {
	return a.tsxStore.Get(handle)
}

func (a *TiledAssets) GetTmxHandle(path file.FilePath) (uint64, bool) {
	return a.tmxStore.GetHandle(path)
}

func (a *TiledAssets) GetTsxHandle(path file.FilePath) (uint64, bool) {
	return a.tsxStore.GetHandle(path)
}

func (t *TiledAssets) BuildTilemap(handle uint64) (*Tilemap, *Tmx) {
	tmx, ok := t.tmxStore.Get(handle)
	if !ok {
		return nil, nil
	}

	tilemap, err := buildTilemap(tmx)
	if err != nil {
		return nil, nil
	}

	t.tilemaps[handle] = tilemap
	return tilemap, tmx
}

func (t *TiledAssets) DeleteTilemap(handle uint64) {
	delete(t.tilemaps, handle)
	if buffers, exists := t.tilemapBuffers[handle]; exists {
		for _, layer := range buffers {
			layer.Deallocate()
		}
		delete(t.tilemapBuffers, handle)
	}
}

func (t *TiledAssets) GetTilemap(handle uint64) (*Tilemap, bool) {
	tilemap, ok := t.tilemaps[handle]
	return tilemap, ok
}

// GetTilemapBuffer returns true if the buffer was resized
func (t *TiledAssets) GetTilemapBuffer(handle uint64, width, height int, layer int) (*ebiten.Image, bool) {
	buffers, exists := t.tilemapBuffers[handle]
	if !exists {
		buffers = make([]*ebiten.Image, 0)
	}

	for len(buffers) <= layer {
		buffers = append(buffers, ebiten.NewImage(width, height))
	}

	t.tilemapBuffers[handle] = buffers
	buffer := buffers[layer]

	if buffer.Bounds().Dx() == width && buffer.Bounds().Dy() == height {
		return buffer, false
	}

	buffer.Deallocate()
	buffer = ebiten.NewImage(width, height)
	buffers[layer] = buffer

	return buffer, true
}

func (a *TiledAssets) ImportAsset(assets engine.Assets, fileSystem fs.FS, path file.FilePath, raw []byte) error {
	switch path.Ext() {
	case ".tmx":
		return a.importTmx(assets, fileSystem, path, raw)
	case ".tsx":
		return a.importTsx(assets, fileSystem, path, raw)
	default:
		return nil
	}
}

func (a *TiledAssets) DeleteAsset(path file.FilePath) bool {
	// TODO - implement this
	return true
}

func (a *TiledAssets) SupportedExtensions() []file.FileExt {
	return a.supportedExtensions
}

func (a *TiledAssets) importTmx(assets engine.Assets, fileSystem fs.FS, path file.FilePath, raw []byte) error {
	if _, exists := a.tmxStore.GetHandle(path); exists {
		return nil
	}

	var tmx Tmx

	err := xml.Unmarshal(raw, &tmx)
	assert.Fatal(err)

	tmx.Handle = a.tmxStore.Insert(path, &tmx)

	if len(tmx.Tilesets) > 0 {
		dir := filepath.Dir(path.String())

		for i := range tmx.Tilesets {
			tileset := &tmx.Tilesets[i]

			if tileset.Source == "" {
				continue
			}

			tsxPath := file.ResolvedPath(dir, tileset.Source)
			tileset.Source = tsxPath.String()

			err := assets.Load(fileSystem, tsxPath)
			assert.Fatal(err)

			handle, exists := a.tsxStore.GetHandle(tsxPath)
			assert.True(exists, fmt.Sprintf("Failed to load TSX asset at path %s", tsxPath))

			tileset.Handle = handle
		}
	}

	return nil
}

func (a *TiledAssets) importTsx(assets engine.Assets, fileSystem fs.FS, path file.FilePath, raw []byte) error {
	if _, exists := a.tsxStore.GetHandle(path); exists {
		return nil
	}

	var tsx Tsx

	err := xml.Unmarshal(raw, &tsx)
	assert.Fatal(err)

	tsx.Handle = a.tsxStore.Insert(path, &tsx)

	dir := filepath.Dir(path.String())

	if tsx.Image.Source == "" {
		return nil
	}

	imgPath := file.ResolvedPath(dir, tsx.Image.Source)
	tsx.Image.Source = imgPath.String()

	err = assets.Load(fileSystem, imgPath)
	assert.Fatal(err)

	imgHandle, exists := a.imageAssets.GetHandle(imgPath)
	assert.True(exists, fmt.Sprintf("Failed to load image asset for TSX at path %s", imgPath))

	tsx.Image.Handle = imgHandle
	a.imageAssets.ExtractUniformFrames(imgHandle, tsx.TileWidth, tsx.TileHeight)

	return nil
}
