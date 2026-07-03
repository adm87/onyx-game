package aseprite

import (
	"github.com/adm87/onyx/pkg/engine"
	"github.com/adm87/onyx/pkg/modules/images"
	"github.com/yohamta/donburi"
)

// moduleID is a unique identifier for the AsepriteModule type.
var moduleID = engine.TypeHash[AsepriteModule]()

func ModuleID() uint64 {
	return moduleID
}

type AsepriteModule interface {
	engine.Module

	Library() *AsepriteLibrary
	Systems() *AsepriteSystems

	CreateSprite(ecs donburi.World, opts ...SpriteOption) *donburi.Entry
}

type module struct {
	library *AsepriteLibrary
	systems *AsepriteSystems

	imageModule images.ImageModule
}

func NewModule() AsepriteModule {
	library := NewAsepriteLibrary()
	systems := NewAsepriteSystems(library)
	return &module{
		library: library,
		systems: systems,
	}
}

func (m *module) OnRegister(game engine.Game) {
	imageModule := engine.GetModule[images.ImageModule](game, images.ModuleID())
	m.imageModule = imageModule
	m.library.imageAssets = imageModule.Assets()
}

func (m *module) ID() uint64 {
	return ModuleID()
}

func (m *module) Library() *AsepriteLibrary {
	return m.library
}

func (m *module) Systems() *AsepriteSystems {
	return m.systems
}

func (m *module) CreateSprite(ecs donburi.World, opts ...SpriteOption) *donburi.Entry {
	options := DefaultSpriteOptions()
	for _, opt := range opts {
		opt(options)
	}

	entry := m.imageModule.CreateImage(ecs,
		images.WithHandle(options.ImageOptions.Handle),
		images.WithAnchor(options.ImageOptions.Anchor.X, options.ImageOptions.Anchor.Y),
		images.WithFrame(options.ImageOptions.Frame),
		images.WithColor(options.ImageOptions.Color),
		images.WithFilter(options.ImageOptions.Filter),
		images.WithTransformOptions(options.ImageOptions.TransformOptions...),
		images.WithRendererOptions(options.ImageOptions.RendererOptions...),
	)

	SetAnimationState(entry, options.State)
	SetAnimator(entry, &AnimatorModel{
		Loops:     options.Loops,
		direction: 1,
	})
	SetClip(entry, options.Clip)

	return entry
}
