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

type ECSRenderAdapter interface {
	PrepareRenderingTasks(
		entry *donburi.Entry,
		renderer *renderer.RendererModel,
		pool *engine.RenderingPool,
		viewport geom.AABB,
		viewMatrix ebiten.GeoM) []*engine.RenderingTask
}

type ECSRenderPipeline struct {
	world donburi.World

	adapters    *slotmap.SlotMap[ECSRenderAdapter]
	partitioner *ECSGrid
	pool        *engine.RenderingPool

	viewport   geom.AABB
	viewMatrix ebiten.GeoM

	tasks []*engine.RenderingTask
}

func NewECSRenderPipeline(world donburi.World, partitioner *ECSGrid) *ECSRenderPipeline {
	return &ECSRenderPipeline{
		world:       world,
		partitioner: partitioner,
		adapters:    slotmap.New[ECSRenderAdapter](0),
		tasks:       make([]*engine.RenderingTask, 0, 100),
	}
}

func (r *ECSRenderPipeline) AddAdapter(adapter ECSRenderAdapter) uint64 {
	return r.adapters.Insert(adapter)
}

func (r *ECSRenderPipeline) GetRenderingTasks(pool *engine.RenderingPool) []*engine.RenderingTask {
	mainCamera, found := camera.GetMainCamera(r.world)
	if !found {
		return nil
	}

	r.viewport, r.viewMatrix = camera.GetView(mainCamera)
	r.tasks = r.tasks[:0]
	r.pool = pool

	r.partitioner.Query(r.viewport, r.getRenderingTasks)

	slices.SortFunc(r.tasks, func(i, j *engine.RenderingTask) int {
		if i.Layer != j.Layer {
			return i.Layer - j.Layer
		}
		if i.ZIndex < j.ZIndex {
			return -1
		} else if i.ZIndex > j.ZIndex {
			return 1
		}
		return 0
	})

	return r.tasks
}

func (r *ECSRenderPipeline) getRenderingTasks(item donburi.Entity) {
	entry := r.world.Entry(item)

	renderer := renderer.GetRenderer(entry)
	if renderer == nil || !renderer.Visible {
		return
	}

	if adapter, exists := r.adapters.Get(renderer.Type); exists {
		tasks := adapter.PrepareRenderingTasks(
			entry,
			renderer,
			r.pool,
			r.viewport,
			r.viewMatrix,
		)
		r.tasks = append(r.tasks, tasks...)
	}
}
