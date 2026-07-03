package debug

import (
	"github.com/adm87/onyx/pkg/engine"
	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/adm87/onyx/pkg/plugins/collision"
	"github.com/adm87/onyx/pkg/plugins/ecs"
	"github.com/hajimehoshi/ebiten/v2"
)

var pluginID = engine.TypeHash[DebugPlugin]()

func PluginID() uint64 {
	return pluginID
}

type DebugPlugin interface {
	engine.Plugin

	Render(target *ebiten.Image, viewport geom.AABB, viewMatrix ebiten.GeoM)

	ToggleCollisionBounds()
	ToggleCollisionInfo()

	ToggleTransformBounds()
	ToggleTransformInfo()

	ToggleRendering()
}

type plugin struct {
	game engine.Game

	ecsPlugin       ecs.ECSPlugin
	collisionPlugin collision.CollisionPlugin

	debugDrawCollisionBounds bool
	debugDrawCollisionInfo   bool

	debugDrawTransformBounds bool
	debugDrawTransformInfo   bool
}

func NewPlugin() DebugPlugin {
	return &plugin{}
}

func (p *plugin) ID() uint64 {
	return PluginID()
}

func (p *plugin) OnRegister(game engine.Game) {
	p.game = game

	p.ecsPlugin = engine.GetPlugin[ecs.ECSPlugin](game, ecs.PluginID())
	p.collisionPlugin = engine.GetPlugin[collision.CollisionPlugin](game, collision.PluginID())
}

func (p *plugin) Render(target *ebiten.Image, viewport geom.AABB, viewMatrix ebiten.GeoM) {
	if p.debugDrawCollisionBounds {
		p.DrawCollisionBounds(target, viewport, viewMatrix)
	}
	if p.debugDrawCollisionInfo {
		p.DrawCollisionInfo(target, viewport, viewMatrix)
	}
	if p.debugDrawTransformBounds {
		p.DrawTransformationBounds(target, viewport, viewMatrix)
	}
	if p.debugDrawTransformInfo {
		p.DrawTransformationInfo(target, viewport, viewMatrix)
	}
}

func (p *plugin) ToggleRendering() {
	if p.game.Renderer().IsEnabled() {
		p.game.Renderer().Disable()
	} else {
		p.game.Renderer().Enable()
	}
}

func (p *plugin) ToggleCollisionBounds() {
	p.debugDrawTransformBounds = false
	p.debugDrawCollisionBounds = !p.debugDrawCollisionBounds
}

func (p *plugin) ToggleCollisionInfo() {
	p.debugDrawTransformInfo = false
	p.debugDrawCollisionInfo = !p.debugDrawCollisionInfo
}

func (p *plugin) ToggleTransformBounds() {
	p.debugDrawCollisionBounds = false
	p.debugDrawTransformBounds = !p.debugDrawTransformBounds
}

func (p *plugin) ToggleTransformInfo() {
	p.debugDrawCollisionInfo = false
	p.debugDrawTransformInfo = !p.debugDrawTransformInfo
}
