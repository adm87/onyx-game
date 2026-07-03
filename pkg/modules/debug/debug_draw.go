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

func drawAABB(target *ebiten.Image, aabb geom.AABB, viewMatrix ebiten.GeoM, color color.RGBA) {
	minX, minY := viewMatrix.Apply(aabb.Min.X, aabb.Min.Y)
	maxX, maxY := viewMatrix.Apply(aabb.Max.X, aabb.Max.Y)
	vector.StrokeRect(target, float32(minX), float32(minY), float32(maxX-minX), float32(maxY-minY), 2, color, false)
}

func drawVec2(target *ebiten.Image, pos geom.Vec2, viewMatrix ebiten.GeoM, color color.RGBA) {
	screenX, screenY := viewMatrix.Apply(pos.X, pos.Y)
	vector.FillRect(target, float32(screenX)-2, float32(screenY)-2, 4, 4, color, false)
}

func drawText(target *ebiten.Image, text string, pos geom.Vec2, viewMatrix ebiten.GeoM) {
	screenX, screenY := viewMatrix.Apply(pos.X, pos.Y)
	ebitenutil.DebugPrintAt(target, text, int(screenX), int(screenY))
}

func (m *module) DrawCollisionBounds(target *ebiten.Image, viewport geom.AABB, viewMatrix ebiten.GeoM) {
	m.collisionModule.StaticQuery(viewport, func(entry *donburi.Entry) {
		drawAABB(target, collision.GetWorldCollider(entry), viewMatrix, staticCollisionBoundsColor)
	})
	m.collisionModule.DynamicQuery(viewport, func(entry *donburi.Entry) {
		drawAABB(target, collision.GetWorldCollider(entry), viewMatrix, dynamicCollisionBoundsColor)
	})
}

func (m *module) DrawTransformationBounds(target *ebiten.Image, viewport geom.AABB, viewMatrix ebiten.GeoM) {
	m.ecsModule.QueryAll(viewport, func(entry *donburi.Entry) {
		if entry.HasComponent(camera.MainCamera) {
			return // Camera will have its own debug system to make sure information is drawn correctly
		}
		drawAABB(target, transform.GetWorldBounds(entry), viewMatrix, transformBoundsColor)
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

		drawText(target, text, bounds.Min, viewMatrix)
		drawVec2(target, geom.Vec2{X: posX, Y: posY}, viewMatrix, transformPositionColor)
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

		drawText(target, text, bounds.Min, viewMatrix)
		drawVec2(target, geom.Vec2{X: posX, Y: posY}, viewMatrix, transformPositionColor)
	})
}
