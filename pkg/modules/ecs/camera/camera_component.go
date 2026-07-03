package camera

import (
	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/adm87/onyx/pkg/modules/ecs/transform"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

var MainCamera = donburi.NewTag("Main Camera")

type CameraOptions struct {
	IsMainCamera     bool
	Zoom             float64
	TransformOptions []transform.TransformOption
}

type CameraOption func(*CameraOptions)

func WithTransformOptions(options ...transform.TransformOption) CameraOption {
	return func(o *CameraOptions) {
		o.TransformOptions = options
	}
}

func WithZoom(zoom float64) CameraOption {
	return func(o *CameraOptions) {
		o.Zoom = zoom
	}
}

func AsMainCamera() CameraOption {
	return func(o *CameraOptions) {
		o.IsMainCamera = true
	}
}

func NewCamera(world donburi.World, options ...CameraOption) *donburi.Entry {
	cameraOptions := &CameraOptions{}
	for _, option := range options {
		option(cameraOptions)
	}

	entry := transform.NewTransform(world, cameraOptions.TransformOptions...)
	if cameraOptions.IsMainCamera {
		entry.AddComponent(MainCamera)
	}

	transform.SetScale(entry, cameraOptions.Zoom, cameraOptions.Zoom)

	return entry
}

func GetMainCamera(world donburi.World) (*donburi.Entry, bool) {
	return MainCamera.First(world)
}

func GetView(entry *donburi.Entry) (viewport geom.AABB, viewMatrix ebiten.GeoM) {
	matrix := transform.GetMatrix(entry)
	bounds := transform.GetBounds(entry)

	minX, minY := matrix.Apply(bounds.Min.X, bounds.Min.Y)
	maxX, maxY := matrix.Apply(bounds.Max.X, bounds.Max.Y)

	matrix.Invert()
	return geom.AABB{
		Min: geom.Vec2{X: minX, Y: minY},
		Max: geom.Vec2{X: maxX, Y: maxY},
	}, matrix
}

func GetZoom(entry *donburi.Entry) float64 {
	x, y := transform.GetScale(entry)
	return (x + y) * 0.5
}

func SetZoom(entry *donburi.Entry, zoom float64) {
	transform.SetScale(entry, zoom, zoom)
}

func ToWorld(entry *donburi.Entry, position geom.Vec2) geom.Vec2 {
	matrix := transform.GetMatrix(entry)
	worldX, worldY := matrix.Apply(position.X, position.Y)
	return geom.Vec2{X: worldX, Y: worldY}
}

func ToScreen(entry *donburi.Entry, position geom.Vec2) geom.Vec2 {
	_, viewMatrix := GetView(entry)
	screenX, screenY := viewMatrix.Apply(position.X, position.Y)
	return geom.Vec2{X: screenX, Y: screenY}
}
