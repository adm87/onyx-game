package camera

import (
	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/adm87/onyx/pkg/modules/ecs/transform"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

var MainCamera = donburi.NewTag("Main Camera")

func GetMainCamera(world donburi.World) (*donburi.Entry, bool) {
	return MainCamera.First(world)
}

func GetView(entry *donburi.Entry) (viewport geom.AABB, viewMatrix ebiten.GeoM) {
	matrix := transform.GetMatrix(entry)
	matrix.Invert()

	invMatrix := matrix
	invMatrix.Invert()

	bounds := transform.GetBounds(entry)
	minX, minY := invMatrix.Apply(bounds.Min.X, bounds.Min.Y)
	maxX, maxY := invMatrix.Apply(bounds.Max.X, bounds.Max.Y)

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
	_, viewMatrix := GetView(entry)
	viewMatrix.Invert()

	worldX, worldY := viewMatrix.Apply(position.X, position.Y)
	return geom.Vec2{X: worldX, Y: worldY}
}

func ToScreen(entry *donburi.Entry, position geom.Vec2) geom.Vec2 {
	_, viewMatrix := GetView(entry)

	screenX, screenY := viewMatrix.Apply(position.X, position.Y)
	return geom.Vec2{X: screenX, Y: screenY}
}
