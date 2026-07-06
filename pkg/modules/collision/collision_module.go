package collision

import (
	"math"

	"github.com/adm87/onyx/pkg/engine"
	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/adm87/onyx/pkg/engine/storage/pool"
	"github.com/adm87/onyx/pkg/modules/ecs"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/events"
)

var (
	StaticCollisionHandled = events.NewEventType[*donburi.Entry]()
)

var moduleID = engine.TypeHash[CollisionModule]()

func ModuleID() uint64 {
	return moduleID
}

type CollisionModule interface {
	engine.Module

	StaticGrid() *ecs.ECSGrid
	DynamicGrid() *ecs.ECSGrid

	QueryAll(area geom.AABB, callback func(entry *donburi.Entry))
	QueryStatic(area geom.AABB, callback func(entry *donburi.Entry))
	QueryDynamic(area geom.AABB, callback func(entry *donburi.Entry))

	QueryAllCollisions(area geom.AABB, callback func(entry *donburi.Entry, collisions []*HitInfo))
	QueryStaticCollisions(area geom.AABB, callback func(entry *donburi.Entry, collisions []*HitInfo))
	QueryDynamicCollisions(area geom.AABB, callback func(entry *donburi.Entry, collisions []*HitInfo))

	UpdateAllCollisions(entry *donburi.Entry)
	UpdateStaticCollisions(entry *donburi.Entry)
	UpdateDynamicCollisions(entry *donburi.Entry)

	GetStaticCollisions(entry *donburi.Entry) ([]*HitInfo, bool)
	GetDynamicCollisions(entry *donburi.Entry) ([]*HitInfo, bool)
}

type module struct {
	world donburi.World

	staticGrid  *ecs.ECSGrid
	dynamicGrid *ecs.ECSGrid

	staticCollisions  map[donburi.Entity]CollisionInfo
	dynamicCollisions map[donburi.Entity]CollisionInfo

	hitPool *pool.Pool[*HitInfo]
}

func NewModule() CollisionModule {
	return &module{
		staticGrid:        ecs.NewEntityGrid(32),
		dynamicGrid:       ecs.NewEntityGrid(32),
		staticCollisions:  make(map[donburi.Entity]CollisionInfo),
		dynamicCollisions: make(map[donburi.Entity]CollisionInfo),
		hitPool: pool.NewPool(
			func() *HitInfo {
				return &HitInfo{}
			},
		),
	}
}

func (m *module) OnRegister(game engine.Game) {
	ecsModule := engine.GetModule[ecs.ECSModule](game, ecs.ModuleID())
	m.world = ecsModule.World()

	ecs.OnAdded.Subscribe(m.world, m.addCollisionEntry)
	ecs.OnRemoved.Subscribe(m.world, m.removeCollisionEntry)
	ecs.OnUpdated.Subscribe(m.world, m.updateCollisionEntry)
}

func (m *module) ID() uint64 {
	return ModuleID()
}

func (m *module) StaticGrid() *ecs.ECSGrid {
	return m.staticGrid
}

func (m *module) DynamicGrid() *ecs.ECSGrid {
	return m.dynamicGrid
}

func (m *module) QueryAll(area geom.AABB, callback func(entry *donburi.Entry)) {
	m.QueryStatic(area, callback)
	m.QueryDynamic(area, callback)
}

func (m *module) QueryStatic(area geom.AABB, callback func(entry *donburi.Entry)) {
	m.staticGrid.Query(area, func(entity donburi.Entity) {
		entry := m.world.Entry(entity)
		callback(entry)
	})
}

func (m *module) QueryDynamic(area geom.AABB, callback func(entry *donburi.Entry)) {
	m.dynamicGrid.Query(area, func(entity donburi.Entity) {
		entry := m.world.Entry(entity)
		callback(entry)
	})
}

func (m *module) QueryAllCollisions(area geom.AABB, callback func(entry *donburi.Entry, collisions []*HitInfo)) {
	m.QueryStaticCollisions(area, callback)
	m.QueryDynamicCollisions(area, callback)
}

func (m *module) QueryStaticCollisions(area geom.AABB, callback func(entry *donburi.Entry, collisions []*HitInfo)) {
	// Note: Query dynamic grid for static collisions since static only dynamic objects can collide with static objects
	m.dynamicGrid.Query(area, func(entity donburi.Entity) {
		entry := m.world.Entry(entity)
		if collisions, found := m.staticCollisions[entity]; found && len(collisions.info) > 0 {
			callback(entry, collisions.info)
		}
	})
}

func (m *module) QueryDynamicCollisions(area geom.AABB, callback func(entry *donburi.Entry, collisions []*HitInfo)) {
	m.dynamicGrid.Query(area, func(entity donburi.Entity) {
		entry := m.world.Entry(entity)
		if collisions, found := m.dynamicCollisions[entity]; found && len(collisions.info) > 0 {
			callback(entry, collisions.info)
		}
	})
}

func (m *module) UpdateAllCollisions(entry *donburi.Entry) {
	m.UpdateStaticCollisions(entry)
	m.UpdateDynamicCollisions(entry)
}

func (m *module) UpdateStaticCollisions(entry *donburi.Entry) {
	entity := entry.Entity()
	if collisions, ok := m.staticCollisions[entity]; ok {
		m.returnHitInfos(collisions.info)
		collisions.info = m.updateCollisions(entry, m.staticGrid, collisions.info[:0])
		m.staticCollisions[entity] = collisions
		if len(collisions.info) > 0 {
			OnStaticCollisions.Publish(m.world, &CollisionEvent{
				Entry: entry,
				Hits:  collisions.info,
			})
			OnStaticCollisions.ProcessEvents(m.world)
		}
	}
}

func (m *module) UpdateDynamicCollisions(entry *donburi.Entry) {
	entity := entry.Entity()
	if collisions, ok := m.dynamicCollisions[entity]; ok {
		m.returnHitInfos(collisions.info)
		collisions.info = m.updateCollisions(entry, m.dynamicGrid, collisions.info[:0])
		m.dynamicCollisions[entity] = collisions
		if len(collisions.info) > 0 {
			OnDynamicCollisions.Publish(m.world, &CollisionEvent{
				Entry: entry,
				Hits:  collisions.info,
			})
			OnDynamicCollisions.ProcessEvents(m.world)
		}
	}
}

func (m *module) GetStaticCollisions(entry *donburi.Entry) ([]*HitInfo, bool) {
	entity := entry.Entity()
	if collisions, ok := m.staticCollisions[entity]; ok && len(collisions.info) > 0 {
		return collisions.info, true
	}
	return nil, false
}

func (m *module) GetDynamicCollisions(entry *donburi.Entry) ([]*HitInfo, bool) {
	entity := entry.Entity()
	if collisions, ok := m.dynamicCollisions[entity]; ok && len(collisions.info) > 0 {
		return collisions.info, true
	}
	return nil, false
}

func (m *module) returnHitInfos(hit []*HitInfo) {
	for _, h := range hit {
		m.hitPool.Put(h)
	}
}

func (m *module) updateCollisions(entry *donburi.Entry, grid *ecs.ECSGrid, infos []*HitInfo) []*HitInfo {
	collider := GetWorldCollider(entry)
	collision := GetCollision(entry)

	if collision.IsStatic {
		return infos // Static objects do not initiate collisions with other objects.
	}
	if !collision.Enabled {
		return infos // Disabled objects do not initiate collisions with other objects.
	}

	grid.Query(collider, func(e donburi.Entity) {
		other := m.world.Entry(e)
		if other == entry {
			return // Skip self
		}

		otherCollider := GetWorldCollider(other)
		otherCollision := GetCollision(other)

		if !otherCollision.Enabled {
			return // Skip disabled objects
		}

		overlap, normal, depth, ok := m.processCollision(collider, otherCollider)
		if !ok {
			return // Skip if no collision response
		}

		hit := m.hitPool.Get()
		hit.Other = other
		hit.ColliderA = collider
		hit.ColliderB = otherCollider
		hit.Overlap = overlap
		hit.Normal = normal
		hit.Depth = depth

		infos = append(infos, hit)
	})

	return infos
}

func (m *module) processCollision(colliderA, colliderB geom.AABB) (geom.AABB, geom.Vec2, float64, bool) {
	xOverlap := math.Min(colliderA.Max.X, colliderB.Max.X) - math.Max(colliderA.Min.X, colliderB.Min.X)
	yOverlap := math.Min(colliderA.Max.Y, colliderB.Max.Y) - math.Max(colliderA.Min.Y, colliderB.Min.Y)

	if xOverlap <= 0 || yOverlap <= 0 {
		return geom.AABB{}, geom.Vec2{}, 0, false // No collision
	}

	overlap := geom.AABB{
		Min: geom.Vec2{X: math.Max(colliderA.Min.X, colliderB.Min.X), Y: math.Max(colliderA.Min.Y, colliderB.Min.Y)},
		Max: geom.Vec2{X: math.Min(colliderA.Max.X, colliderB.Max.X), Y: math.Min(colliderA.Max.Y, colliderB.Max.Y)},
	}

	centerA := colliderA.Center()
	centerB := colliderB.Center()

	if xOverlap < yOverlap {
		if centerA.X < centerB.X {
			return overlap, geom.Vec2{X: -1, Y: 0}, xOverlap, true // Collision from the left
		} else {
			return overlap, geom.Vec2{X: 1, Y: 0}, xOverlap, true // Collision from the right
		}
	} else {
		if centerA.Y < centerB.Y {
			return overlap, geom.Vec2{X: 0, Y: -1}, yOverlap, true // Collision from above
		} else {
			return overlap, geom.Vec2{X: 0, Y: 1}, yOverlap, true // Collision from below
		}
	}
}

func (m *module) addCollisionEntry(_ donburi.World, entry *donburi.Entry) {
	if !entry.HasComponent(Collision) {
		return
	}
	m.addEntry(entry)
}

func (m *module) removeCollisionEntry(_ donburi.World, entry *donburi.Entry) {
	if !entry.HasComponent(Collision) {
		return
	}
	m.removeEntry(entry)
}

func (m *module) updateCollisionEntry(_ donburi.World, entry *donburi.Entry) {
	if !entry.HasComponent(Collision) {
		return
	}
	m.updateEntry(entry)
}

func (m *module) addEntry(entry *donburi.Entry) {
	index := GetCollisionIndex(entry)
	if index > 0 {
		return // Already indexed
	}

	entity := entry.Entity()
	m.staticCollisions[entity] = CollisionInfo{}
	m.dynamicCollisions[entity] = CollisionInfo{}

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

	entity := entry.Entity()
	delete(m.staticCollisions, entity)
	delete(m.dynamicCollisions, entity)

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
