package collision

import (
	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/adm87/onyx/pkg/plugins/ecs/transform"
	"github.com/yohamta/donburi"
)

type CollisionLayer uint8

func (l CollisionLayer) IsDefault() bool {
	return l == 0
}

type CollisionOptions struct {
	Layer    CollisionLayer
	Collider *geom.AABB
	Enabled  bool
	IsStatic bool
}

type CollisionOption func(*CollisionOptions)

type CollisionModel struct {
	Layer    CollisionLayer
	Enabled  bool
	IsStatic bool
}

var (
	Collision = donburi.NewComponentType[CollisionModel]()
	Collider  = donburi.NewComponentType[geom.AABB]()
)

func defaultCollisionOptions() *CollisionOptions {
	return &CollisionOptions{
		Layer:    0,
		Enabled:  true,
		IsStatic: false,
	}
}

func WithCollisionLayer(layer CollisionLayer) CollisionOption {
	return func(o *CollisionOptions) {
		o.Layer = layer
	}
}

func WithCollider(min, max geom.Vec2) CollisionOption {
	return func(o *CollisionOptions) {
		o.Collider = &geom.AABB{
			Min: min,
			Max: max,
		}
	}
}

func AsStatic() CollisionOption {
	return func(o *CollisionOptions) {
		o.IsStatic = true
	}
}

func AddCollisionComponent(entry *donburi.Entry, options ...CollisionOption) {
	opts := defaultCollisionOptions()
	for _, opt := range options {
		opt(opts)
	}

	donburi.Add(entry, Collision, &CollisionModel{
		Layer:    opts.Layer,
		Enabled:  opts.Enabled,
		IsStatic: opts.IsStatic,
	})

	// If no collider is provided, use the bounds of the transform component as the collider
	if opts.Collider == nil {
		bounds := transform.GetBounds(entry)
		donburi.Add(entry, Collider, &bounds)
	} else {
		donburi.Add(entry, Collider, opts.Collider)
	}
}

func GetCollision(entry *donburi.Entry) *CollisionModel {
	if !entry.HasComponent(Collision) {
		return &CollisionModel{}
	}
	return Collision.Get(entry)
}

func GetCollider(entry *donburi.Entry) geom.AABB {
	if !entry.HasComponent(Collider) {
		return geom.AABB{}
	}
	return *Collider.Get(entry)
}

func GetWorldCollider(entry *donburi.Entry) geom.AABB {
	collider := GetCollider(entry)
	matrix := transform.GetMatrix(entry)

	x1, y1 := matrix.Apply(collider.Min.X, collider.Min.Y)
	x2, y2 := matrix.Apply(collider.Max.X, collider.Max.Y)

	return geom.AABB{
		Min: geom.Vec2{
			X: min(x1, x2),
			Y: min(y1, y2),
		},
		Max: geom.Vec2{
			X: max(x1, x2),
			Y: max(y1, y2),
		},
	}
}
