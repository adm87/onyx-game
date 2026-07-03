package aseprite

import (
	"encoding/json"
	"image"

	"github.com/adm87/onyx/pkg/modules/images"
)

type AsepriteLibrary struct {
	imageAssets *images.ImageAssets
	animations  map[uint64]*AnimationData
}

func NewAsepriteLibrary() *AsepriteLibrary {
	return &AsepriteLibrary{
		animations: make(map[uint64]*AnimationData),
	}
}

func (l *AsepriteLibrary) BuildAnimations(imgHandle uint64, data []byte) (*AnimationData, error) {
	var animations *AnimationData

	err := json.Unmarshal(data, &animations)
	if err != nil {
		return nil, err
	}

	animations.Meta.Clips = make(map[string]FrameTag, len(animations.Meta.FrameTags))
	for i := range animations.Meta.FrameTags {
		tag := animations.Meta.FrameTags[i]
		animations.Meta.Clips[tag.Name] = tag
	}

	rects := make([]image.Rectangle, len(animations.Frames))
	for i, frame := range animations.Frames {
		rects[i] = image.Rect(
			frame.Frame.X,
			frame.Frame.Y,
			frame.Frame.X+frame.Frame.W,
			frame.Frame.Y+frame.Frame.H,
		)
	}

	l.imageAssets.ExtractFrames(imgHandle, rects)
	l.animations[imgHandle] = animations

	return animations, nil
}

func (l *AsepriteLibrary) DeleteAnimations(handle uint64) {
	delete(l.animations, handle)
}
