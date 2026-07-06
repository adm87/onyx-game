package ecs

import (
	"github.com/adm87/onyx/pkg/engine"
	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/adm87/onyx/pkg/modules/ecs/transform"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/events"
)

var moduleID = engine.TypeHash[ECSModule]()

func ModuleID() uint64 {
	return moduleID
}

var (
	OnAdded   = events.NewEventType[*donburi.Entry]()
	OnRemoved = events.NewEventType[*donburi.Entry]()
	OnUpdated = events.NewEventType[*donburi.Entry]()
)

type ECSModule interface {
	engine.Module

	Add(entries ...*donburi.Entry)
	Remove(entries ...*donburi.Entry)
	Update(entries ...*donburi.Entry)

	ProcessPendingUpdates()

	QueryAll(area geom.AABB, fn func(entry *donburi.Entry))
	QueryResolution(area geom.AABB, fn func(entry *donburi.Entry))

	World() donburi.World
	RenderPipeline() *ECSRenderPipeline
}

type module struct {
	world donburi.World

	renderPipeline *ECSRenderPipeline
	grid           *ECSGrid

	pendingUpdates map[donburi.Entity]uint64
	updatePass     uint64
}

func NewModule() ECSModule {
	world := donburi.NewWorld()
	ecsGrid := NewEntityGrid(32, 64, 128, 256, 512)
	renderPipeline := NewECSRenderPipeline(world, ecsGrid)
	return &module{
		world:          world,
		renderPipeline: renderPipeline,
		grid:           ecsGrid,
		pendingUpdates: make(map[donburi.Entity]uint64),
		updatePass:     1,
	}
}

func (m *module) OnRegister(game engine.Game) {

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

func (m *module) ProcessPendingUpdates() {
	m.processUpdates()
	m.updatePass++

	OnUpdated.ProcessEvents(m.world)
}

// processUpdates updates entities within the ecs grid that have been marked for updates, and queues update events for those entities for other systems.
func (m *module) processUpdates() {
	for entity, pass := range m.pendingUpdates {
		if pass == m.updatePass {
			entry := m.world.Entry(entity)
			m.updateEntry(entry)

			OnUpdated.Publish(m.world, entry)
			delete(m.pendingUpdates, entity)
		}
	}
}

func (m *module) Add(entries ...*donburi.Entry) {
	for i := range entries {
		entry := entries[i]

		OnAdded.Publish(m.world, entry)
		m.addEntry(entry)
	}
	OnAdded.ProcessEvents(m.world)
}

func (m *module) Remove(entries ...*donburi.Entry) {
	for i := range entries {
		entry := entries[i]

		OnRemoved.Publish(m.world, entry)
		OnRemoved.ProcessEvents(m.world) // Process events immediately to ensure that any systems listening for removal can react before the entity is removed from the world.

		m.removeEntry(entry)
	}
}

func (m *module) Update(entries ...*donburi.Entry) {
	for i := range entries {
		entry := entries[i]
		entity := entry.Entity()
		if m.pendingUpdates[entity] != m.updatePass {
			m.pendingUpdates[entity] = m.updatePass
		}
	}
}

func (m *module) addEntry(entry *donburi.Entry) {
	transform.SetIndex(entry, m.grid.Insert(entry.Entity(), transform.GetWorldBounds(entry)))
}

func (m *module) removeEntry(entry *donburi.Entry) {
	entity := entry.Entity()
	m.grid.Remove(entity)
	m.world.Remove(entity)
	delete(m.pendingUpdates, entity)
}

func (m *module) updateEntry(entry *donburi.Entry) {
	transform.SetIndex(entry, m.grid.Update(entry.Entity(), transform.GetWorldBounds(entry)))
}

func (m *module) QueryAll(area geom.AABB, fn func(entry *donburi.Entry)) {
	m.processUpdates() // Ensure no stale data is used during the query

	m.grid.Query(area, func(entity donburi.Entity) {
		entry := m.world.Entry(entity)
		fn(entry)
	})
}

func (m *module) QueryResolution(area geom.AABB, fn func(entry *donburi.Entry)) {
	m.processUpdates() // Ensure no stale data is used during the query

	partition, _ := m.grid.NearestGrid(area)
	partition.Query(area, func(entity donburi.Entity) {
		entry := m.world.Entry(entity)
		fn(entry)
	})
}
