package collision

import (
	"github.com/adm87/onyx/pkg/engine"
	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/adm87/onyx/pkg/modules/ecs"
	"github.com/yohamta/donburi"
)

var moduleID = engine.TypeHash[CollisionModule]()

func ModuleID() uint64 {
	return moduleID
}

type CollisionModule interface {
	engine.Module

	QueryAll(area geom.AABB, callback func(entry *donburi.Entry))
	StaticQuery(area geom.AABB, callback func(entry *donburi.Entry))
	DynamicQuery(area geom.AABB, callback func(entry *donburi.Entry))
}

type module struct {
	ecsModule ecs.ECSModule

	staticGrid  *ecs.ECSGrid
	dynamicGrid *ecs.ECSGrid
}

func NewModule() CollisionModule {
	return &module{
		staticGrid:  ecs.NewEntityGrid(32),
		dynamicGrid: ecs.NewEntityGrid(32),
	}
}

func (m *module) OnRegister(game engine.Game) {
	m.ecsModule = engine.GetModule[ecs.ECSModule](game, ecs.ModuleID())
	m.ecsModule.AddECSCallbacks(ecs.ECSCallbacks{
		Added:   m.addCollisionEntries,
		Removed: m.removeCollisionEntries,
		Updated: m.updateCollisionEntries,
	})
}

func (m *module) ID() uint64 {
	return ModuleID()
}

func (m *module) QueryAll(area geom.AABB, callback func(entry *donburi.Entry)) {
	m.StaticQuery(area, callback)
	m.DynamicQuery(area, callback)
}

func (m *module) StaticQuery(area geom.AABB, callback func(entry *donburi.Entry)) {
	m.staticGrid.Query(area, func(entity donburi.Entity) {
		entry := m.ecsModule.World().Entry(entity)
		callback(entry)
	})
}

func (m *module) DynamicQuery(area geom.AABB, callback func(entry *donburi.Entry)) {
	m.dynamicGrid.Query(area, func(entity donburi.Entity) {
		entry := m.ecsModule.World().Entry(entity)
		callback(entry)
	})
}

func (m *module) addCollisionEntries(entries []*donburi.Entry) {
	for i := range entries {
		entry := entries[i]
		if !entry.HasComponent(Collision) {
			continue
		}
		m.addEntry(entry)
	}
}

func (m *module) removeCollisionEntries(entries []*donburi.Entry) {
	for i := range entries {
		entry := entries[i]
		if !entry.HasComponent(Collision) {
			continue
		}
		m.removeEntry(entry)
	}
}

func (m *module) updateCollisionEntries(entries []*donburi.Entry) {
	for i := range entries {
		entry := entries[i]
		if !entry.HasComponent(Collision) {
			continue
		}
		m.updateEntry(entry)
	}
}

func (m *module) addEntry(entry *donburi.Entry) {
	index := GetCollisionIndex(entry)
	if index > 0 {
		return // Already indexed
	}

	collision := GetCollision(entry)
	collider := GetWorldCollider(entry)

	if collision.IsStatic {
		index = m.staticGrid.Insert(entry.Entity(), collider)
	} else {
		index = m.dynamicGrid.Insert(entry.Entity(), collider)
	}

	SetCollisionIndex(entry, index)
}

func (m *module) removeEntry(entry *donburi.Entry) {
	index := GetCollisionIndex(entry)

	if index == 0 {
		return // Not indexed
	}

	collision := GetCollision(entry)

	if collision.IsStatic {
		m.staticGrid.Remove(entry.Entity())
	} else {
		m.dynamicGrid.Remove(entry.Entity())
	}

	SetCollisionIndex(entry, 0)
}

func (m *module) updateEntry(entry *donburi.Entry) {
	index := GetCollisionIndex(entry)
	if index == 0 {
		m.addEntry(entry)
		return // Not indexed, so add it
	}

	collision := GetCollision(entry)
	collider := GetWorldCollider(entry)

	if collision.IsStatic {
		index = m.staticGrid.Update(entry.Entity(), collider)
	} else {
		index = m.dynamicGrid.Update(entry.Entity(), collider)
	}

	SetCollisionIndex(entry, index)
}
