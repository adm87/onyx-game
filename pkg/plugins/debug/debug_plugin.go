package debug

import (
	"fmt"
	"image/color"

	"github.com/adm87/onyx/pkg/engine"
	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/adm87/onyx/pkg/plugins/ecs"
	"github.com/adm87/onyx/pkg/plugins/ecs/transform"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/yohamta/donburi"
)

var pluginID = engine.TypeHash[DebugPlugin]()

func PluginID() uint64 {
	return pluginID
}

type DebugPlugin interface {
	engine.Plugin

	PathTransformBounds(viewport geom.AABB, viewMatrix ebiten.GeoM)

	ResetPath()

	DrawPath(target *ebiten.Image, color color.RGBA)
	DrawTransformationInfo(target *ebiten.Image, viewport geom.AABB, viewMatrix ebiten.GeoM)
}

type plugin struct {
	path vector.Path

	ecsPlugin ecs.ECSPlugin
}

func NewPlugin() DebugPlugin {
	return &plugin{}
}

func (p *plugin) ID() uint64 {
	return PluginID()
}

func (p *plugin) OnRegister(game engine.Game) {
	p.ecsPlugin = engine.GetPlugin[ecs.ECSPlugin](game, ecs.PluginID())
}

func (p *plugin) ResetPath() {
	p.path.Reset()
}

func (p *plugin) PathTransformBounds(viewport geom.AABB, viewMatrix ebiten.GeoM) {
	p.ecsPlugin.QueryAll(viewport, func(entry *donburi.Entry) {
		bounds := transform.GetWorldBounds(entry)

		minX, minY := viewMatrix.Apply(bounds.Min.X, bounds.Min.Y)
		maxX, maxY := viewMatrix.Apply(bounds.Max.X, bounds.Max.Y)

		p.path.MoveTo(float32(minX), float32(minY))
		p.path.LineTo(float32(maxX), float32(minY))
		p.path.LineTo(float32(maxX), float32(maxY))
		p.path.LineTo(float32(minX), float32(maxY))
		p.path.Close()
	})
}

func (p *plugin) PathColliderBounds(viewport geom.AABB, viewMatrix ebiten.GeoM) {

}

func (p *plugin) DrawPath(target *ebiten.Image, color color.RGBA) {
	strokeOpts := &vector.StrokeOptions{Width: 2}
	drawOpts := &vector.DrawPathOptions{}

	drawOpts.ColorScale.ScaleWithColor(color)
	vector.StrokePath(target, &p.path, strokeOpts, drawOpts)

}

func (p *plugin) DrawTransformationInfo(target *ebiten.Image, viewport geom.AABB, viewMatrix ebiten.GeoM) {
	p.ecsPlugin.QueryAll(viewport, func(entry *donburi.Entry) {
		bounds := transform.GetWorldBounds(entry)
		posX, posY := transform.GetPosition(entry)
		scaleX, scaleY := transform.GetScale(entry)
		rotation := transform.GetRotation(entry)

		text := fmt.Sprintf("Entity ID: %d\n", entry.Entity())
		text += fmt.Sprintf("Position: (%.2f, %.2f)\n", posX, posY)
		text += fmt.Sprintf("Scale: (%.2f, %.2f)\n", scaleX, scaleY)
		text += fmt.Sprintf("Rotation: %.2f\n", rotation)
		text += fmt.Sprintf("Bounds: Min(%.2f, %.2f), Max(%.2f, %.2f)\n", bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Max.Y)

		screenX, screenY := viewMatrix.Apply(bounds.Min.X, bounds.Min.Y)
		ebitenutil.DebugPrintAt(target, text, int(screenX), int(screenY))

		screenX, screenY = viewMatrix.Apply(posX, posY)
		vector.FillRect(target, float32(screenX)-2, float32(screenY)-2, 4, 4, color.RGBA{G: 255, A: 255}, false)
	})
}
