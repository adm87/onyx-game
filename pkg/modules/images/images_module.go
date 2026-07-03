package images

import (
	"github.com/adm87/onyx/pkg/engine"
	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/adm87/onyx/pkg/modules/ecs"
	"github.com/adm87/onyx/pkg/modules/ecs/renderer"
	"github.com/adm87/onyx/pkg/modules/ecs/transform"
	"github.com/yohamta/donburi"
)

var moduleID = engine.TypeHash[ImageModule]()

func ModuleID() uint64 {
	return moduleID
}

type ImageModule interface {
	engine.Module

	Assets() *ImageAssets
	CreateImage(world donburi.World, opts ...ImageOption) *donburi.Entry
}

type module struct {
	assets   *ImageAssets
	renderer *ImageECSRenderer

	rendererType uint64
}

func NewModule() ImageModule {
	assets := NewImageAssets()
	renderer := NewImageECSRenderer(assets)
	return &module{
		assets:   assets,
		renderer: renderer,
	}
}

func (m *module) OnRegister(game engine.Game) {
	game.Assets().AddAdapter(m.assets)

	ecsModule := engine.GetModule[ecs.ECSModule](game, ecs.ModuleID())
	m.rendererType = ecsModule.RenderPipeline().AddAdapter(m.renderer)
}

func (m *module) ID() uint64 {
	return ModuleID()
}

func (m *module) Assets() *ImageAssets {
	return m.assets
}

func (m *module) CreateImage(world donburi.World, opts ...ImageOption) *donburi.Entry {
	options := DefaultImageOptions()
	for _, opt := range opts {
		opt(options)
	}

	var bounds geom.AABB

	if img, exists := m.assets.GetFrame(options.Handle, options.Frame); exists {
		anchor := options.Anchor

		width, height := img.Bounds().Dx(), img.Bounds().Dy()
		bounds.Min = geom.Vec2{
			X: -anchor.X * float64(width),
			Y: -anchor.Y * float64(height),
		}
		bounds.Max = geom.Vec2{
			X: bounds.Min.X + float64(width),
			Y: bounds.Min.Y + float64(height),
		}
	}

	options.TransformOptions = append(options.TransformOptions, transform.WithBounds(bounds.Min, bounds.Max))
	options.RendererOptions = append(options.RendererOptions, renderer.WithRendererType(m.rendererType))

	entry := transform.NewTransform(world, options.TransformOptions...)
	renderer.AddRenderer(entry, options.RendererOptions...)

	donburi.Add(entry, Image, &ImageModel{
		Handle: options.Handle,
		Frame:  options.Frame,
		Anchor: options.Anchor,
		Color:  options.Color,
		Filter: options.Filter,
	})

	return entry
}
