package renderer

import "github.com/yohamta/donburi"

type RenderOptions struct {
	RendererType uint64
	Layer        int
	ZIndex       float32
	Visible      bool
}

type RenderOption func(*RenderOptions)

type RendererModel struct {
	Type    uint64
	Layer   int
	ZIndex  float32
	Visible bool
}

var Renderer = donburi.NewComponentType[RendererModel]()

func defaultRendererOptions() *RenderOptions {
	return &RenderOptions{
		RendererType: 0,
		Layer:        0,
		ZIndex:       0,
		Visible:      true,
	}
}

func WithRendererType(rendererType uint64) RenderOption {
	return func(opts *RenderOptions) {
		opts.RendererType = rendererType
	}
}

func WithLayer(layer int) RenderOption {
	return func(opts *RenderOptions) {
		opts.Layer = layer
	}
}

func WithVisibility(visible bool) RenderOption {
	return func(opts *RenderOptions) {
		opts.Visible = visible
	}
}

func WithZIndex(zIndex float32) RenderOption {
	return func(opts *RenderOptions) {
		opts.ZIndex = zIndex
	}
}

func NewRenderer(world donburi.World, opts ...RenderOption) *donburi.Entry {
	return AddRenderer(world.Entry(world.Create(Renderer)), opts...)
}

func AddRenderer(entry *donburi.Entry, options ...RenderOption) *donburi.Entry {
	SetRenderer(entry, options...)
	return entry
}

func GetRenderer(entry *donburi.Entry) *RendererModel {
	if !entry.HasComponent(Renderer) {
		return nil
	}
	return Renderer.Get(entry)
}

func SetRenderer(entry *donburi.Entry, options ...RenderOption) {
	opts := defaultRendererOptions()
	for _, opt := range options {
		opt(opts)
	}
	donburi.Add(entry, Renderer, &RendererModel{
		Type:    opts.RendererType,
		Layer:   opts.Layer,
		ZIndex:  opts.ZIndex,
		Visible: opts.Visible,
	})
}

func GetLayer(entry *donburi.Entry) int {
	if !entry.HasComponent(Renderer) {
		return 0
	}
	return Renderer.Get(entry).Layer
}

func SetLayer(entry *donburi.Entry, layer int) {
	if !entry.HasComponent(Renderer) {
		return
	}
	renderer := Renderer.Get(entry)
	renderer.Layer = layer
}

func GetZIndex(entry *donburi.Entry) float32 {
	if !entry.HasComponent(Renderer) {
		return 0
	}
	return Renderer.Get(entry).ZIndex
}

func SetZIndex(entry *donburi.Entry, zIndex float32) {
	if !entry.HasComponent(Renderer) {
		return
	}
	renderer := Renderer.Get(entry)
	renderer.ZIndex = zIndex
}

func SetVisibility(entry *donburi.Entry, visible bool) {
	if !entry.HasComponent(Renderer) {
		return
	}
	renderer := Renderer.Get(entry)
	renderer.Visible = visible
}

func IsVisible(entry *donburi.Entry) bool {
	if !entry.HasComponent(Renderer) {
		return false
	}
	return Renderer.Get(entry).Visible
}
