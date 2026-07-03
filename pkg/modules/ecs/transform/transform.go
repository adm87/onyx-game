package transform

import (
	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

type TransformOption func(*TransformOptions)

type TransformOptions struct {
	X, Y      float64
	OriginX   float64
	OriginY   float64
	ScaleX    float64
	ScaleY    float64
	Rotation  float64
	Index     uint64
	BoundsMin geom.Vec2
	BoundsMax geom.Vec2
}

type TransformModel struct {
	x, y    float64
	ox, oy  float64
	sx, sy  float64
	r       float64
	isDirty bool
}

var (
	Transform       = donburi.NewComponentType[TransformModel]()
	TransformMatrix = donburi.NewComponentType[ebiten.GeoM]()
	TransformBounds = donburi.NewComponentType[geom.AABB]()
	TransformIndex  = donburi.NewComponentType[uint64]()
)

func defaultTransformOptions() *TransformOptions {
	return &TransformOptions{
		X:        0,
		Y:        0,
		OriginX:  0,
		OriginY:  0,
		ScaleX:   1,
		ScaleY:   1,
		Rotation: 0,
	}
}

func WithOrigin(originX, originY float64) TransformOption {
	return func(o *TransformOptions) {
		o.OriginX = originX
		o.OriginY = originY
	}
}

func WithPosition(x, y float64) TransformOption {
	return func(o *TransformOptions) {
		o.X = x
		o.Y = y
	}
}

func WithScale(scaleX, scaleY float64) TransformOption {
	return func(o *TransformOptions) {
		o.ScaleX = scaleX
		o.ScaleY = scaleY
	}
}

func WithRotation(rotation float64) TransformOption {
	return func(o *TransformOptions) {
		o.Rotation = rotation
	}
}

func WithBounds(min, max geom.Vec2) TransformOption {
	return func(o *TransformOptions) {
		o.BoundsMin = min
		o.BoundsMax = max
	}
}

func WithIndex(index uint64) TransformOption {
	return func(o *TransformOptions) {
		o.Index = index
	}
}

func NewTransform(world donburi.World, opts ...TransformOption) *donburi.Entry {
	return AddTransform(world.Entry(world.Create(
		Transform,
		TransformMatrix,
		TransformBounds,
		TransformIndex,
	)), opts...)
}

func AddTransform(entry *donburi.Entry, opts ...TransformOption) *donburi.Entry {
	options := defaultTransformOptions()
	for _, opt := range opts {
		opt(options)
	}

	SetTransform(entry, &TransformModel{
		x:       options.X,
		y:       options.Y,
		sx:      options.ScaleX,
		sy:      options.ScaleY,
		r:       options.Rotation,
		ox:      options.OriginX,
		oy:      options.OriginY,
		isDirty: true,
	})

	SetBounds(entry, &geom.AABB{
		Min: options.BoundsMin,
		Max: options.BoundsMax,
	})

	SetIndex(entry,
		options.Index,
	)

	entry.AddComponent(TransformMatrix)

	return entry
}

func GetTransform(entry *donburi.Entry) *TransformModel {
	if !entry.HasComponent(Transform) {
		return &TransformModel{}
	}
	return Transform.Get(entry)
}

func SetTransform(entry *donburi.Entry, t *TransformModel) {
	t.isDirty = true
	donburi.Add(entry, Transform, t)
}

func GetBounds(entry *donburi.Entry) geom.AABB {
	if !entry.HasComponent(TransformBounds) {
		return geom.AABB{}
	}
	return *TransformBounds.Get(entry)
}

func SetBounds(entry *donburi.Entry, bounds *geom.AABB) {
	donburi.Add(entry, TransformBounds, bounds)
}

func GetWorldBounds(entry *donburi.Entry) geom.AABB {
	bounds := GetBounds(entry)
	matrix := GetMatrix(entry)

	x1, y1 := matrix.Apply(bounds.Min.X, bounds.Min.Y)
	x2, y2 := matrix.Apply(bounds.Max.X, bounds.Max.Y)

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

func GetMatrix(entry *donburi.Entry) ebiten.GeoM {
	t := Transform.Get(entry)
	m := TransformMatrix.Get(entry)

	if t.isDirty {
		m.Reset()
		m.Translate(-t.ox, -t.oy)
		m.Scale(t.sx, t.sy)
		m.Rotate(t.r)
		m.Translate(t.x, t.y)
		t.isDirty = false
	}

	return *m
}

func GetIndex(entry *donburi.Entry) uint64 {
	if !entry.HasComponent(TransformIndex) {
		return 0
	}
	return *TransformIndex.Get(entry)
}

func SetIndex(entry *donburi.Entry, index uint64) {
	donburi.Add(entry, TransformIndex, &index)
}

func GetOrigin(entry *donburi.Entry) (float64, float64) {
	if !entry.HasComponent(Transform) {
		return 0, 0
	}
	t := Transform.Get(entry)
	return t.ox, t.oy
}

func SetOrigin(entry *donburi.Entry, ox, oy float64) {
	if !entry.HasComponent(Transform) {
		return
	}

	t := Transform.Get(entry)
	if t.ox == ox && t.oy == oy {
		return
	}

	t.ox = ox
	t.oy = oy
	t.isDirty = true
}

func GetPosition(entry *donburi.Entry) (float64, float64) {
	if !entry.HasComponent(Transform) {
		return 0, 0
	}
	t := Transform.Get(entry)
	return t.x, t.y
}

func SetPosition(entry *donburi.Entry, x, y float64) {
	if !entry.HasComponent(Transform) {
		return
	}

	t := Transform.Get(entry)
	if t.x == x && t.y == y {
		return
	}

	t.x = x
	t.y = y
	t.isDirty = true
}

func GetScale(entry *donburi.Entry) (float64, float64) {
	if !entry.HasComponent(Transform) {
		return 1, 1
	}
	t := Transform.Get(entry)
	return t.sx, t.sy
}

func SetScale(entry *donburi.Entry, sx, sy float64) {
	if !entry.HasComponent(Transform) {
		return
	}

	t := Transform.Get(entry)
	if t.sx == sx && t.sy == sy {
		return
	}

	t.sx = sx
	t.sy = sy
	t.isDirty = true
}

func GetRotation(entry *donburi.Entry) float64 {
	if !entry.HasComponent(Transform) {
		return 0
	}
	return Transform.Get(entry).r
}

func SetRotation(entry *donburi.Entry, r float64) {
	if !entry.HasComponent(Transform) {
		return
	}

	t := Transform.Get(entry)
	if t.r == r {
		return
	}

	t.r = r
	t.isDirty = true
}

func IsDirty(entry *donburi.Entry) bool {
	if !entry.HasComponent(Transform) {
		return false
	}
	return Transform.Get(entry).isDirty
}

func Translate(entry *donburi.Entry, dx, dy float64) {
	if !entry.HasComponent(Transform) || (dx == 0 && dy == 0) {
		return
	}
	t := Transform.Get(entry)
	t.x += dx
	t.y += dy
	t.isDirty = true
}

func Scale(entry *donburi.Entry, sx, sy float64) {
	if !entry.HasComponent(Transform) || (sx == 1 && sy == 1) {
		return
	}
	t := Transform.Get(entry)
	t.sx *= sx
	t.sy *= sy
	t.isDirty = true
}

func Rotate(entry *donburi.Entry, dr float64) {
	if !entry.HasComponent(Transform) || dr == 0 {
		return
	}
	t := Transform.Get(entry)
	t.r += dr
	t.isDirty = true
}
