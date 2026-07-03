package engine

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Renderer interface {
	IsEnabled() bool
	Enable()
	Disable()
	SetRenderPipeline(RenderPipeline)
	SetClearColor(color.Color)
}

type RenderPipeline interface {
	Run(target *ebiten.Image)
}

type RenderingTask struct {
	Buffer  *ebiten.Image
	Options *ebiten.DrawImageOptions
	Layer   int
	ZIndex  float32
}

type renderer struct {
	enabled bool

	screen *screen
	logger *logger

	color color.Color

	pipeline RenderPipeline
}

func newRenderer(screen *screen, logger *logger) *renderer {
	return &renderer{
		screen:  screen,
		logger:  logger,
		enabled: true,
	}
}

func (r *renderer) IsEnabled() bool {
	return r.enabled
}

func (r *renderer) Enable() {
	r.enabled = true
}

func (r *renderer) Disable() {
	r.enabled = false
}

func (r *renderer) SetRenderPipeline(p RenderPipeline) {
	r.pipeline = p
}

func (r *renderer) SetClearColor(color color.Color) {
	r.color = color
}

func (r *renderer) render(target *ebiten.Image) {
	if r.color == nil {
		target.Clear()
	} else {
		target.Fill(r.color)
	}

	if !r.enabled {
		return
	}

	if r.pipeline == nil {
		ebitenutil.DebugPrint(target, "No render pipeline set")
		return
	}

	r.pipeline.Run(target)
}
