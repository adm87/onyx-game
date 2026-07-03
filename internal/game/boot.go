package game

import (
	"context"
	"path/filepath"

	"github.com/adm87/onyx/content"
	"github.com/adm87/onyx/internal/game/cli"
	"github.com/adm87/onyx/internal/game/onyx"
	"github.com/adm87/onyx/pkg/engine"
	"github.com/adm87/onyx/pkg/engine/assert"
	"github.com/adm87/onyx/pkg/modules/aseprite"
	"github.com/adm87/onyx/pkg/modules/collision"
	"github.com/adm87/onyx/pkg/modules/debug"
	"github.com/adm87/onyx/pkg/modules/ecs"
	"github.com/adm87/onyx/pkg/modules/images"
	"github.com/adm87/onyx/pkg/modules/tiled"
	"github.com/hajimehoshi/ebiten/v2"
)

func Boot() error {
	args, err := cli.ParseArgs()
	assert.Fatal(err)

	path, err := filepath.Abs(args.RootDir)
	assert.Fatal(err)

	content.InitContentDirectories(path)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	game := engine.NewGame(
		engine.WithTitle("Onyx"),
		engine.WithScreenSize(1280, 720),
		engine.WithScreenScale(engine.ScreenScaleFill),
		engine.WithFullscreen(args.Fullscreen),
		engine.WithInitialScene(onyx.GameplaySceneID),
		engine.WithFilter(ebiten.FilterNearest),
		engine.WithModules(
			aseprite.NewModule(),
			collision.NewModule(),
			debug.NewModule(),
			ecs.NewModule(),
			images.NewModule(),
			tiled.NewModule(),
		),
	).WithContext(ctx)

	ecsModule := engine.GetModule[ecs.ECSModule](game, ecs.ModuleID())

	renderer := game.Renderer()
	renderer.SetRenderPipeline(ecsModule.RenderPipeline())

	return onyx.NewGame(game).Start()
}
