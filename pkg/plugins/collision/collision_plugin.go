package collision

import (
	"github.com/adm87/onyx/pkg/engine"
	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/adm87/onyx/pkg/plugins/ecs"
	"github.com/yohamta/donburi"
)

var pluginID = engine.TypeHash[CollisionPlugin]()

func PluginID() uint64 {
	return pluginID
}

type CollisionPlugin interface {
	engine.Plugin

	QueryAll(area geom.AABB, callback func(entry *donburi.Entry))
	StaticQuery(area geom.AABB, callback func(entry *donburi.Entry))
	DynamicQuery(area geom.AABB, callback func(entry *donburi.Entry))
}

type plugin struct {
	ecsPlugin ecs.ECSPlugin

	staticGrid  *ecs.ECSGrid
	dynamicGrid *ecs.ECSGrid
}

func NewPlugin() CollisionPlugin {
	return &plugin{
		staticGrid:  ecs.NewEntityGrid(32),
		dynamicGrid: ecs.NewEntityGrid(32),
	}
}

func (p *plugin) OnRegister(game engine.Game) {
	p.ecsPlugin = engine.GetPlugin[ecs.ECSPlugin](game, ecs.PluginID())
	p.ecsPlugin.AddECSCallbacks(ecs.ECSCallbacks{
		Added:   p.addCollisionEntries,
		Removed: p.removeCollisionEntries,
		Updated: p.updateCollisionEntries,
	})
}

func (p *plugin) ID() uint64 {
	return PluginID()
}

func (p *plugin) QueryAll(area geom.AABB, callback func(entry *donburi.Entry)) {
	p.StaticQuery(area, callback)
	p.DynamicQuery(area, callback)
}

func (p *plugin) StaticQuery(area geom.AABB, callback func(entry *donburi.Entry)) {
	p.staticGrid.Query(area, func(entity donburi.Entity) {
		entry := p.ecsPlugin.World().Entry(entity)
		callback(entry)
	})
}

func (p *plugin) DynamicQuery(area geom.AABB, callback func(entry *donburi.Entry)) {
	p.dynamicGrid.Query(area, func(entity donburi.Entity) {
		entry := p.ecsPlugin.World().Entry(entity)
		callback(entry)
	})
}

func (p *plugin) addCollisionEntries(entries []*donburi.Entry) {
	for i := range entries {
		entry := entries[i]
		if !entry.HasComponent(Collision) {
			continue
		}
		p.addEntry(entry)
	}
}

func (p *plugin) removeCollisionEntries(entries []*donburi.Entry) {
	for i := range entries {
		entry := entries[i]
		if !entry.HasComponent(Collision) {
			continue
		}
		p.removeEntry(entry)
	}
}

func (p *plugin) updateCollisionEntries(entries []*donburi.Entry) {
	for i := range entries {
		entry := entries[i]
		if !entry.HasComponent(Collision) {
			continue
		}
		p.updateEntry(entry)
	}
}

func (p *plugin) addEntry(entry *donburi.Entry) {
	index := GetCollisionIndex(entry)
	if index > 0 {
		return // Already indexed
	}

	collision := GetCollision(entry)
	collider := GetWorldCollider(entry)

	if collision.IsStatic {
		index = p.staticGrid.Insert(entry.Entity(), collider)
	} else {
		index = p.dynamicGrid.Insert(entry.Entity(), collider)
	}

	SetCollisionIndex(entry, index)
}

func (p *plugin) removeEntry(entry *donburi.Entry) {
	index := GetCollisionIndex(entry)

	if index == 0 {
		return // Not indexed
	}

	collision := GetCollision(entry)

	if collision.IsStatic {
		p.staticGrid.Remove(entry.Entity())
	} else {
		p.dynamicGrid.Remove(entry.Entity())
	}

	SetCollisionIndex(entry, 0)
}

func (p *plugin) updateEntry(entry *donburi.Entry) {
	index := GetCollisionIndex(entry)
	if index == 0 {
		p.addEntry(entry)
		return // Not indexed, so add it
	}

	collision := GetCollision(entry)
	collider := GetWorldCollider(entry)

	if collision.IsStatic {
		index = p.staticGrid.Update(entry.Entity(), collider)
	} else {
		index = p.dynamicGrid.Update(entry.Entity(), collider)
	}

	SetCollisionIndex(entry, index)
}
