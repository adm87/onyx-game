package tiled

import (
	"github.com/adm87/onyx/pkg/engine"
	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/adm87/onyx/pkg/modules/ecs"
	"github.com/adm87/onyx/pkg/modules/ecs/renderer"
	"github.com/adm87/onyx/pkg/modules/ecs/transform"
	"github.com/adm87/onyx/pkg/modules/images"
	"github.com/yohamta/donburi"
)

var moduleID = engine.TypeHash[TiledModule]()

func ModuleID() uint64 {
	return moduleID
}

type TiledModule interface {
	engine.Module

	Assets() *TiledAssets
	CreateTilemap(world donburi.World, opts ...TilemapOption) *donburi.Entry
}

type module struct {
	assets   *TiledAssets
	renderer *TiledECSRenderer

	rendererType uint64
}

func NewModule() TiledModule {
	assets := NewTiledAssets()
	renderer := NewTiledECSRenderer(assets)
	return &module{
		assets:   assets,
		renderer: renderer,
	}
}

func (m *module) OnRegister(game engine.Game) {
	game.Assets().AddAdapter(m.assets)

	ecsModule := engine.GetModule[ecs.ECSModule](game, ecs.ModuleID())
	m.rendererType = ecsModule.RenderPipeline().AddAdapter(m.renderer)

	imageModule := engine.GetModule[images.ImageModule](game, images.ModuleID())
	m.assets.imageAssets = imageModule.Assets()
	m.renderer.imageAssets = imageModule.Assets()

	m.renderer.screen = game.Screen()
}

func (m *module) ID() uint64 {
	return ModuleID()
}

func (m *module) Assets() *TiledAssets {
	return m.assets
}

func (m *module) CreateTilemap(world donburi.World, opts ...TilemapOption) *donburi.Entry {
	entry := NewTilemap(world, opts...)

	var bounds geom.AABB

	tilemapHandle := GetTilemapHandle(entry)
	if tilemap, exists := m.assets.GetTilemap(tilemapHandle); exists {
		bounds = tilemap.Bounds()
	}

	transform.AddTransform(entry,
		transform.WithBounds(bounds.Min, bounds.Max),
	)

	renderer.AddRenderer(entry,
		renderer.WithRendererType(m.rendererType),
	)

	return entry
}
