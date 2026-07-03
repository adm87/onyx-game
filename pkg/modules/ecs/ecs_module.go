package ecs

import (
	"github.com/adm87/onyx/pkg/engine"
	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/adm87/onyx/pkg/modules/ecs/transform"
	"github.com/yohamta/donburi"
)

var moduleID = engine.TypeHash[ECSModule]()

func ModuleID() uint64 {
	return moduleID
}

type ECSCallbacks struct {
	Added   func(entries []*donburi.Entry)
	Removed func(entries []*donburi.Entry)
	Updated func(entries []*donburi.Entry)
}

type ECSModule interface {
	engine.Module

	AddECSCallbacks(callbacks ECSCallbacks)

	Add(entries ...*donburi.Entry)
	Remove(entries ...*donburi.Entry)
	Update(entries ...*donburi.Entry)

	QueryAll(area geom.AABB, fn func(entry *donburi.Entry))
	QueryResolution(area geom.AABB, fn func(entry *donburi.Entry))

	World() donburi.World
	RenderPipeline() *ECSRenderPipeline
}

type module struct {
	world donburi.World

	renderPipeline *ECSRenderPipeline
	grid           *ECSGrid

	callbacks []ECSCallbacks
}

func NewModule() ECSModule {
	world := donburi.NewWorld()
	ecsGrid := NewEntityGrid(32, 64, 128, 256, 512)
	renderPipeline := NewECSRenderPipeline(world, ecsGrid)
	return &module{
		world:          world,
		renderPipeline: renderPipeline,
		grid:           ecsGrid,
		callbacks:      make([]ECSCallbacks, 0),
	}
}

func (m *module) OnRegister(game engine.Game) {
	m.AddECSCallbacks(ECSCallbacks{
		Added:   m.addEntries,
		Removed: m.removeEntries,
		Updated: m.updateEntries,
	})
}

func (m *module) ID() uint64 {
	return ModuleID()
}

func (m *module) RenderPipeline() *ECSRenderPipeline {
	return m.renderPipeline
}

func (m *module) World() donburi.World {
	return m.world
}

func (m *module) AddECSCallbacks(callbacks ECSCallbacks) {
	m.callbacks = append(m.callbacks, callbacks)
}

func (m *module) Add(entries ...*donburi.Entry) {
	for i := range m.callbacks {
		callbacks := m.callbacks[i]
		if callbacks.Added != nil {
			callbacks.Added(entries)
		}
	}
}

func (m *module) Remove(entries ...*donburi.Entry) {
	for i := range m.callbacks {
		callbacks := m.callbacks[i]
		if callbacks.Removed != nil {
			callbacks.Removed(entries)
		}
	}
}

func (m *module) Update(entries ...*donburi.Entry) {
	for i := range m.callbacks {
		callbacks := m.callbacks[i]
		if callbacks.Updated != nil {
			callbacks.Updated(entries)
		}
	}
}

func (m *module) addEntries(entries []*donburi.Entry) {
	for i := range entries {
		entry := entries[i]
		transform.SetIndex(entry, m.grid.Insert(entry.Entity(), transform.GetWorldBounds(entry)))
	}
}

func (m *module) removeEntries(entries []*donburi.Entry) {
	for i := range entries {
		entity := entries[i].Entity()
		m.grid.Remove(entity)
		m.world.Remove(entity)
	}
}

func (m *module) updateEntries(entries []*donburi.Entry) {
	for i := range entries {
		entry := entries[i]
		transform.SetIndex(entry, m.grid.Update(entry.Entity(), transform.GetWorldBounds(entry)))
	}
}

func (m *module) QueryAll(area geom.AABB, fn func(entry *donburi.Entry)) {
	m.grid.Query(area, func(entity donburi.Entity) {
		entry := m.world.Entry(entity)
		fn(entry)
	})
}

func (m *module) QueryResolution(area geom.AABB, fn func(entry *donburi.Entry)) {
	partition, _ := m.grid.nearestGrid(area)
	partition.Query(area, func(entity donburi.Entity) {
		entry := m.world.Entry(entity)
		fn(entry)
	})
}
