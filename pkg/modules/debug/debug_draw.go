package debug

import (
	"fmt"
	"image/color"

	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/adm87/onyx/pkg/modules/collision"
	"github.com/adm87/onyx/pkg/modules/ecs/camera"
	"github.com/adm87/onyx/pkg/modules/ecs/transform"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/yohamta/donburi"
)

var (
	staticCollisionBoundsColor  = color.RGBA{R: 255, A: 255}
	dynamicCollisionBoundsColor = color.RGBA{R: 255, G: 255, A: 255}

	transformBoundsColor   = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	transformPositionColor = color.RGBA{G: 255, A: 255}
)

func (m *module) DrawCollisionBounds(target *ebiten.Image, viewport geom.AABB, viewMatrix ebiten.GeoM) {
	m.collisionModule.StaticQuery(viewport, func(entry *donburi.Entry) {
		bounds := collision.GetWorldCollider(entry)

		minX, minY := viewMatrix.Apply(bounds.Min.X, bounds.Min.Y)
		maxX, maxY := viewMatrix.Apply(bounds.Max.X, bounds.Max.Y)

		vector.StrokeRect(target, float32(minX), float32(minY), float32(maxX-minX), float32(maxY-minY), 2,
			staticCollisionBoundsColor, false)
	})
	m.collisionModule.DynamicQuery(viewport, func(entry *donburi.Entry) {
		bounds := collision.GetWorldCollider(entry)

		minX, minY := viewMatrix.Apply(bounds.Min.X, bounds.Min.Y)
		maxX, maxY := viewMatrix.Apply(bounds.Max.X, bounds.Max.Y)

		vector.StrokeRect(target, float32(minX), float32(minY), float32(maxX-minX), float32(maxY-minY), 2,
			dynamicCollisionBoundsColor, false)
	})
}

func (m *module) DrawTransformationBounds(target *ebiten.Image, viewport geom.AABB, viewMatrix ebiten.GeoM) {
	m.ecsModule.QueryAll(viewport, func(entry *donburi.Entry) {
		if entry.HasComponent(camera.MainCamera) {
			return // Camera will have its own debug system to make sure information is drawn correctly
		}

		bounds := transform.GetWorldBounds(entry)

		minX, minY := viewMatrix.Apply(bounds.Min.X, bounds.Min.Y)
		maxX, maxY := viewMatrix.Apply(bounds.Max.X, bounds.Max.Y)

		vector.StrokeRect(target, float32(minX), float32(minY), float32(maxX-minX), float32(maxY-minY), 2,
			transformBoundsColor, false)
	})
}

func (m *module) DrawCollisionInfo(target *ebiten.Image, viewport geom.AABB, viewMatrix ebiten.GeoM) {
	m.collisionModule.QueryAll(viewport, func(entry *donburi.Entry) {
		col := collision.GetCollision(entry)
		bounds := collision.GetWorldCollider(entry)
		posX, posY := transform.GetPosition(entry)

		text := "Collision Info:\n"
		text += fmt.Sprintf(". Entity ID: %d\n", entry.Entity())
		text += fmt.Sprintf(". Position: (%.2f, %.2f)\n", posX, posY)
		text += fmt.Sprintf(". Collider: Min(%.2f, %.2f), Max(%.2f, %.2f)\n", bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Max.Y)
		text += fmt.Sprintf(". Layer: %d\n", col.Layer)
		text += fmt.Sprintf(". Enabled: %t\n", col.Enabled)
		text += fmt.Sprintf(". IsStatic: %t\n", col.IsStatic)

		screenX, screenY := viewMatrix.Apply(bounds.Min.X, bounds.Min.Y)
		ebitenutil.DebugPrintAt(target, text, int(screenX), int(screenY))

		screenX, screenY = viewMatrix.Apply(posX, posY)
		vector.FillRect(target, float32(screenX)-2, float32(screenY)-2, 4, 4,
			transformPositionColor, false)
	})
}

func (m *module) DrawTransformationInfo(target *ebiten.Image, viewport geom.AABB, viewMatrix ebiten.GeoM) {
	m.ecsModule.QueryAll(viewport, func(entry *donburi.Entry) {
		if entry.HasComponent(camera.MainCamera) {
			return // Camera will have its own debug system to make sure information is drawn correctly
		}

		bounds := transform.GetWorldBounds(entry)
		posX, posY := transform.GetPosition(entry)
		scaleX, scaleY := transform.GetScale(entry)
		rotation := transform.GetRotation(entry)

		text := "Transformation Info:\n"
		text += fmt.Sprintf(". Entity ID: %d\n", entry.Entity())
		text += fmt.Sprintf(". Position: (%.2f, %.2f)\n", posX, posY)
		text += fmt.Sprintf(". Scale: (%.2f, %.2f)\n", scaleX, scaleY)
		text += fmt.Sprintf(". Rotation: %.2f\n", rotation)
		text += fmt.Sprintf(". Bounds: Min(%.2f, %.2f), Max(%.2f, %.2f)\n", bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Max.Y)

		screenX, screenY := viewMatrix.Apply(bounds.Min.X, bounds.Min.Y)
		ebitenutil.DebugPrintAt(target, text, int(screenX), int(screenY))

		screenX, screenY = viewMatrix.Apply(posX, posY)
		vector.FillRect(target, float32(screenX)-2, float32(screenY)-2, 4, 4,
			transformPositionColor, false)
	})
}
