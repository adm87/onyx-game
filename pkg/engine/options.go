package engine

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type Options struct {
	Title           string
	Width           int
	Height          int
	FPS             int
	Fullscreen      bool
	ScreenScale     ScreenScaleMode
	Filter          ebiten.Filter
	BackgroundColor color.RGBA
	InitialScene    SceneID
	Modules         []Module
}

type Option func(*Options)

func WithFullscreen(fullscreen bool) Option {
	return func(c *Options) {
		c.Fullscreen = fullscreen
	}
}

func WithFilter(filter ebiten.Filter) Option {
	return func(c *Options) {
		c.Filter = filter
	}
}

func WithScreenScale(mode ScreenScaleMode) Option {
	return func(c *Options) {
		c.ScreenScale = mode
	}
}

func WithTitle(title string) Option {
	return func(c *Options) {
		c.Title = title
	}
}

func WithScreenSize(width, height int) Option {
	return func(c *Options) {
		c.Width = width
		c.Height = height
	}
}

func WithFPS(fps int) Option {
	return func(c *Options) {
		c.FPS = fps
	}
}

func WithBackgroundColor(color color.RGBA) Option {
	return func(c *Options) {
		c.BackgroundColor = color
	}
}

func WithInitialScene(id SceneID) Option {
	return func(c *Options) {
		c.InitialScene = id
	}
}

func WithModules(modules ...Module) Option {
	return func(c *Options) {
		c.Modules = append(c.Modules, modules...)
	}
}

func defaultConfig() *Options {
	return &Options{
		Title:       "Untitled",
		Width:       800,
		Height:      600,
		FPS:         60,
		Fullscreen:  false,
		ScreenScale: ScreenScaleNone,
		Filter:      ebiten.FilterLinear,
		BackgroundColor: color.RGBA{
			R: 0,
			G: 0,
			B: 0,
			A: 255,
		},
		InitialScene: SceneIDNone,
		Modules:      []Module{},
	}
}

func applyOptions(opts ...Option) *Options {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}
