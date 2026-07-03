package images

import (
	"bytes"
	"image"
	"io/fs"
	"math"

	"github.com/adm87/onyx/pkg/engine"
	"github.com/adm87/onyx/pkg/engine/file"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type ImageAssets struct {
	cache               file.FileStore[*ebiten.Image]
	frames              map[uint64][]*ebiten.Image
	supportedExtensions []file.FileExt
}

func NewImageAssets() *ImageAssets {
	return &ImageAssets{
		cache:               file.NewFileStore[*ebiten.Image](0),
		frames:              make(map[uint64][]*ebiten.Image),
		supportedExtensions: []file.FileExt{".png", ".jpg", ".jpeg"},
	}
}

func (i *ImageAssets) SupportedExtensions() []file.FileExt {
	return i.supportedExtensions
}

func (i *ImageAssets) ImportAsset(_ engine.Assets, _ fs.FS, path file.FilePath, raw []byte) error {
	img, _, err := ebitenutil.NewImageFromReader(bytes.NewReader(raw))
	if err != nil {
		return err
	}
	i.cache.Insert(path, img)
	return nil
}

func (i *ImageAssets) DeleteAsset(path file.FilePath) bool {
	handle, exists := i.cache.GetHandle(path)
	if !exists {
		return false
	}

	if img, deleted := i.cache.Delete(handle); deleted {
		img.Deallocate()
	}

	delete(i.frames, handle)
	return true
}

func (a *ImageAssets) ExtractUniformFrames(handle uint64, frameWidth, frameHeight int) {
	img, exists := a.cache.Get(handle)
	if !exists {
		return
	}

	frames, exists := a.frames[handle]
	if exists {
		frames = frames[:0]
	}

	columns := int(math.Ceil(float64(img.Bounds().Dx()) / float64(frameWidth)))
	rows := int(math.Ceil(float64(img.Bounds().Dy()) / float64(frameHeight)))

	for y := range rows {
		for x := range columns {
			frame := img.SubImage(image.Rect(
				x*frameWidth,
				y*frameHeight,
				(x+1)*frameWidth,
				(y+1)*frameHeight,
			)).(*ebiten.Image)
			frames = append(frames, frame)
		}
	}

	a.frames[handle] = frames
}

func (a *ImageAssets) ExtractFrames(handle uint64, rects []image.Rectangle) {
	img, exists := a.cache.Get(handle)
	if !exists {
		return
	}

	frames, exists := a.frames[handle]
	if exists {
		frames = frames[:0]
	}

	for _, frameRect := range rects {
		frame := img.SubImage(frameRect).(*ebiten.Image)
		frames = append(frames, frame)
	}

	a.frames[handle] = frames
}

func (a *ImageAssets) Get(handle uint64) (*ebiten.Image, bool) {
	img, exists := a.cache.Get(handle)
	if !exists {
		return nil, false
	}
	return img, true
}

func (a *ImageAssets) GetHandle(path file.FilePath) (uint64, bool) {
	handle, exists := a.cache.GetHandle(path)
	if !exists {
		return 0, false
	}
	return handle, true
}

func (a *ImageAssets) GetFrame(handle uint64, index int) (*ebiten.Image, bool) {
	frames, exists := a.frames[handle]
	if !exists || index < 0 || index >= len(frames) {
		return a.Get(handle)
	}
	return frames[index], true
}

func (a *ImageAssets) GetImageSize(handle uint64) (int, int, bool) {
	img, exists := a.cache.Get(handle)
	if !exists {
		return 0, 0, false
	}
	bounds := img.Bounds()
	return bounds.Dx(), bounds.Dy(), true
}

func (a *ImageAssets) GetFrameSize(handle uint64) (int, int, bool) {
	frames, exists := a.frames[handle]
	if !exists || len(frames) == 0 {
		return a.GetImageSize(handle)
	}
	bounds := frames[0].Bounds()
	return bounds.Dx(), bounds.Dy(), true
}
