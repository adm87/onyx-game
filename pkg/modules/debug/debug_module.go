package debug

import (
	"github.com/adm87/onyx/pkg/engine"
	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/adm87/onyx/pkg/modules/collision"
	"github.com/adm87/onyx/pkg/modules/ecs"
	"github.com/hajimehoshi/ebiten/v2"
)

var moduleID = engine.TypeHash[DebugModule]()

func ModuleID() uint64 {
	return moduleID
}

type DebugModule interface {
	engine.Module

	Render(target *ebiten.Image, viewport geom.AABB, viewMatrix ebiten.GeoM)

	ToggleCollisionBounds()
	ToggleCollisionInfo()
	ToggleCollisions()

	ToggleTransformBounds()
	ToggleTransformInfo()

	ToggleRendering()
}

type module struct {
	game engine.Game

	ecsModule       ecs.ECSModule
	collisionModule collision.CollisionModule

	debugDrawCollisionBounds bool
	debugDrawCollisionInfo   bool
	debugDrawCollisions      bool

	debugDrawTransformBounds bool
	debugDrawTransformInfo   bool
}

func NewModule() DebugModule {
	return &module{}
}

func (m *module) ID() uint64 {
	return ModuleID()
}

func (m *module) OnRegister(game engine.Game) {
	m.game = game

	m.ecsModule = engine.GetModule[ecs.ECSModule](game, ecs.ModuleID())
	m.collisionModule = engine.GetModule[collision.CollisionModule](game, collision.ModuleID())
}

func (m *module) Render(target *ebiten.Image, viewport geom.AABB, viewMatrix ebiten.GeoM) {
	if m.debugDrawCollisionBounds {
		m.DrawCollisionBounds(target, viewport, viewMatrix)
	}
	if m.debugDrawCollisionInfo {
		m.DrawCollisionInfo(target, viewport, viewMatrix)
	}
	if m.debugDrawCollisions {
		m.DrawCollisions(target, viewport, viewMatrix)
	}
	if m.debugDrawTransformBounds {
		m.DrawTransformationBounds(target, viewport, viewMatrix)
	}
	if m.debugDrawTransformInfo {
		m.DrawTransformationInfo(target, viewport, viewMatrix)
	}
}

func (m *module) ToggleRendering() {
	if m.game.Renderer().IsEnabled() {
		m.game.Renderer().Disable()
	} else {
		m.game.Renderer().Enable()
	}
}

func (m *module) ToggleCollisionBounds() {
	m.debugDrawTransformBounds = false
	m.debugDrawCollisionBounds = !m.debugDrawCollisionBounds
}

func (m *module) ToggleCollisionInfo() {
	m.debugDrawTransformInfo = false
	m.debugDrawCollisionInfo = !m.debugDrawCollisionInfo
}

func (m *module) ToggleCollisions() {
	m.debugDrawCollisions = !m.debugDrawCollisions
}

func (m *module) ToggleTransformBounds() {
	m.debugDrawCollisionBounds = false
	m.debugDrawTransformBounds = !m.debugDrawTransformBounds
}

func (m *module) ToggleTransformInfo() {
	m.debugDrawCollisionInfo = false
	m.debugDrawTransformInfo = !m.debugDrawTransformInfo
}
