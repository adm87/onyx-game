package images

import (
	"github.com/adm87/onyx/pkg/engine"
	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/adm87/onyx/pkg/modules/ecs/renderer"
	"github.com/adm87/onyx/pkg/modules/ecs/transform"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

type ImageECSRenderer struct {
	imageAssets *ImageAssets
	tasks       []*engine.RenderingTask
}

func NewImageECSRenderer(imageAssets *ImageAssets) *ImageECSRenderer {
	return &ImageECSRenderer{
		imageAssets: imageAssets,
		tasks:       make([]*engine.RenderingTask, 0, 10),
	}
}

func (r *ImageECSRenderer) PrepareRenderingTasks(
	entry *donburi.Entry,
	renderer *renderer.RendererModel,
	pool *engine.RenderingPool,
	viewport geom.AABB,
	viewMatrix ebiten.GeoM) []*engine.RenderingTask {
	r.tasks = r.tasks[:0]

	imgHandle := GetHandle(entry)
	frameIdx := GetFrame(entry)

	if img, exists := r.imageAssets.GetFrame(imgHandle, frameIdx); exists {
		width, height := img.Bounds().Dx(), img.Bounds().Dy()

		anchor := GetAnchor(entry)
		color := GetColor(entry)
		filter := GetFilter(entry)

		matrix := transform.GetMatrix(entry)
		scaleX, scaleY := transform.GetScale(entry)

		aX := anchor.X * float64(width) * scaleX
		aY := anchor.Y * float64(height) * scaleY

		matrix.Translate(-aX, -aY)
		matrix.Concat(viewMatrix)

		task := pool.Get()
		task.Buffer = img
		task.Options = &ebiten.DrawImageOptions{
			GeoM:   matrix,
			Filter: filter,
		}
		task.Options.ColorScale.ScaleWithColor(color)
		task.Options.ColorScale.ScaleAlpha(float32(color.A) / 255)
		task.Layer = renderer.Layer
		task.ZIndex = renderer.ZIndex

		r.tasks = append(r.tasks, task)
	}

	return r.tasks
}
