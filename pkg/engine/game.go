package engine

import (
	"context"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game interface {
	Start() error
	WithContext(ctx context.Context) Game

	Assets() Assets
	Logger() Logger
	Renderer() Renderer
	Scenes() Scenes
	Screen() Screen
	Time() Time
	Modules() Modules
}

type game struct {
	ctx context.Context

	assets   *assets
	logger   *logger
	renderer *renderer
	scenes   *scenes
	screen   *screen
	time     *time

	modules *modules
}

func setupWindow(title string, width, height int) {
	ebiten.SetWindowTitle(title)
	ebiten.SetWindowSize(width, height)
}

func NewGame(opts ...Option) Game {
	cfg := applyOptions(opts...)

	setupWindow(cfg.Title, cfg.Width, cfg.Height)

	logger := newLogger(os.Stdout)

	assets := newAssets(
		logger,
	)

	screen := newScreen(
		cfg.Width,
		cfg.Height,
		cfg.ScreenScale,
		cfg.Filter,
		logger,
	)

	renderer := newRenderer(
		screen,
		logger,
	)

	scenes := newScenes(
		cfg.InitialScene,
		logger,
	)

	time := newTime(
		cfg.FPS,
	)

	m := newModules()

	g := &game{
		ctx:      context.Background(),
		assets:   assets,
		logger:   logger,
		screen:   screen,
		scenes:   scenes,
		time:     time,
		renderer: renderer,
		modules:  m,
	}

	for _, module := range cfg.Modules {
		m.add(module)
	}
	m.Register(g)

	return g
}

func (g *game) Assets() Assets {
	return g.assets
}

func (g *game) Logger() Logger {
	return g.logger
}

func (g *game) Renderer() Renderer {
	return g.renderer
}

func (g *game) Scenes() Scenes {
	return g.scenes
}

func (g *game) Screen() Screen {
	return g.screen
}

func (g *game) Time() Time {
	return g.time
}

func (g *game) Modules() Modules {
	return g.modules
}

func (g *game) WithContext(ctx context.Context) Game {
	if g.ctx == nil {
		return g
	}
	g.ctx = ctx
	return g
}

func (g *game) Start() error {
	return ebiten.RunGame(g)
}

func (g *game) Update() error {
	select {
	case <-g.ctx.Done():
		return g.ctx.Err()
	default:
		g.time.tick()

		dt := g.time.deltaTime.Seconds()
		fdt := g.time.fixedDeltaTime.Seconds()
		steps := g.time.fixedSteps

		return g.scenes.update(steps, dt, fdt)
	}
}

func (g *game) Draw(screen *ebiten.Image) {
	select {
	case <-g.ctx.Done():
		return
	default:
		g.screen.buffer.Clear()

		g.renderer.render(g.screen.buffer)
		g.scenes.render(g.screen.buffer)

		screen.DrawImage(g.screen.buffer, g.screen.options)
	}
}

func (g *game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return g.screen.Layout(outsideWidth, outsideHeight)
}
