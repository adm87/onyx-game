package ecs

import (
	"slices"

	"github.com/adm87/onyx/pkg/engine"
	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/adm87/onyx/pkg/engine/storage/slotmap"
	"github.com/adm87/onyx/pkg/modules/ecs/camera"
	"github.com/adm87/onyx/pkg/modules/ecs/renderer"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

type RenderingPool struct {
	pool []*engine.RenderingTask
	i    int
}

func (m *RenderingPool) Get() *engine.RenderingTask {
	if m.i >= len(m.pool) {
		m.pool = append(m.pool, &engine.RenderingTask{})
	}
	task := m.pool[m.i]
	task.Buffer = nil
	task.Options = nil
	m.i++
	return task
}

type ECSRenderAdapter interface {
	GetRenderingTasks(
		entry *donburi.Entry,
		renderer *renderer.RendererModel,
		viewport geom.AABB,
		viewMatrix ebiten.GeoM,
		pool *RenderingPool,
		bucket []*engine.RenderingTask) []*engine.RenderingTask
}

type ECSRenderPipeline struct {
	world donburi.World

	adapters    *slotmap.SlotMap[ECSRenderAdapter]
	partitioner *ECSGrid

	pool  *RenderingPool
	tasks []*engine.RenderingTask

	cachedViewport   geom.AABB
	cachedViewMatrix ebiten.GeoM
}

func NewECSRenderPipeline(world donburi.World, partitioner *ECSGrid) *ECSRenderPipeline {
	return &ECSRenderPipeline{
		world:       world,
		partitioner: partitioner,
		adapters:    slotmap.New[ECSRenderAdapter](0),
		pool: &RenderingPool{
			pool: make([]*engine.RenderingTask, 0, 100),
		},
		tasks: make([]*engine.RenderingTask, 0, 100),
	}
}

func (r *ECSRenderPipeline) AddAdapter(adapter ECSRenderAdapter) uint64 {
	return r.adapters.Insert(adapter)
}

func (r *ECSRenderPipeline) Run(target *ebiten.Image) {
	mainCamera, found := camera.GetMainCamera(r.world)
	if !found {
		return
	}

	r.tasks = r.tasks[:0]

	r.cachedViewport, r.cachedViewMatrix = camera.GetView(mainCamera)
	r.partitioner.Query(r.cachedViewport, r.renderTaskCollector)

	slices.SortFunc(r.tasks, r.zIndexComparator)
	for i := range r.tasks {
		target.DrawImage(r.tasks[i].Buffer, r.tasks[i].Options)
	}
}

func (r *ECSRenderPipeline) renderTaskCollector(entity donburi.Entity) {
	entry := r.world.Entry(entity)

	renderer := renderer.GetRenderer(entry)
	if renderer == nil || !renderer.Visible {
		return
	}

	if adapter, exists := r.adapters.Get(renderer.Type); exists {
		r.tasks = adapter.GetRenderingTasks(entry, renderer, r.cachedViewport, r.cachedViewMatrix, r.pool, r.tasks)
	}
}

func (r *ECSRenderPipeline) zIndexComparator(a, b *engine.RenderingTask) int {
	if a.Layer != b.Layer {
		return a.Layer - b.Layer
	}
	if a.ZIndex < b.ZIndex {
		return -1
	} else if a.ZIndex > b.ZIndex {
		return 1
	}
	return 0
}
