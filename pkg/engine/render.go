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
	SetBackgroundColor(color.RGBA)
}

type RenderPipeline interface {
	GetRenderingTasks(taskPool *RenderingPool) []*RenderingTask
}

type RenderingTask struct {
	Buffer  *ebiten.Image
	Options *ebiten.DrawImageOptions
	Layer   int
	ZIndex  float32
}

type RenderingPool struct {
	pool []*RenderingTask
	i    int
}

func (p *RenderingPool) Get() *RenderingTask {
	if p.i >= len(p.pool) {
		p.pool = append(p.pool, &RenderingTask{})
	}
	task := p.pool[p.i]
	task.Buffer = nil
	task.Options = nil
	p.i++
	return task
}

type renderer struct {
	enabled bool

	screen *screen
	logger *logger

	pool  *RenderingPool
	color color.RGBA
	tasks []*RenderingTask

	pipeline RenderPipeline
}

func newRenderer(screen *screen, logger *logger) *renderer {
	return &renderer{
		screen: screen,
		logger: logger,
		pool: &RenderingPool{
			pool: make([]*RenderingTask, 0, 100),
		},
		tasks:   make([]*RenderingTask, 0, 100),
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

func (r *renderer) SetBackgroundColor(color color.RGBA) {
	r.color = color
}

func (r *renderer) render(target *ebiten.Image) {
	if !r.enabled {
		return
	}

	if r.pipeline == nil {
		ebitenutil.DebugPrint(target, "No render pipeline set")
		return
	}

	target.Fill(r.color)

	r.tasks = r.tasks[:0]
	r.pool.i = 0

	r.tasks = append(r.tasks, r.pipeline.GetRenderingTasks(r.pool)...)
	for i := range r.tasks {
		target.DrawImage(r.tasks[i].Buffer, r.tasks[i].Options)
	}
}
