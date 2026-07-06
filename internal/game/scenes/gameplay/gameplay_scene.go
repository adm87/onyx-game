package gameplay

import (
	"image/color"
	"time"

	"github.com/adm87/onyx/content"
	"github.com/adm87/onyx/internal/game/movement"
	"github.com/adm87/onyx/internal/game/player"
	"github.com/adm87/onyx/pkg/engine"
	"github.com/adm87/onyx/pkg/engine/file"
	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/adm87/onyx/pkg/modules/aseprite"
	"github.com/adm87/onyx/pkg/modules/collision"
	"github.com/adm87/onyx/pkg/modules/debug"
	"github.com/adm87/onyx/pkg/modules/ecs"
	"github.com/adm87/onyx/pkg/modules/ecs/camera"
	"github.com/adm87/onyx/pkg/modules/ecs/renderer"
	"github.com/adm87/onyx/pkg/modules/ecs/transform"
	"github.com/adm87/onyx/pkg/modules/images"
	"github.com/adm87/onyx/pkg/modules/tiled"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"
)

var transformQuery = donburi.NewQuery(
	filter.Contains(
		transform.Transform,
	),
)

var gameplayManifest = []file.FilePath{
	content.AssetsAsepriteCaptainImg,
	content.AssetsAsepriteCaptainJson,
	content.AssetsTiledGym04,
}

type Scene struct {
	game engine.Game

	tilemapEntry *donburi.Entry
	spriteEntry  *donburi.Entry
	cameraEntry  *donburi.Entry

	asprite   aseprite.AsepriteModule
	collision collision.CollisionModule
	debug     debug.DebugModule
	ecs       ecs.ECSModule
}

func NewScene(game engine.Game) *Scene {
	return &Scene{
		game:      game,
		asprite:   engine.GetModule[aseprite.AsepriteModule](game, aseprite.ModuleID()),
		collision: engine.GetModule[collision.CollisionModule](game, collision.ModuleID()),
		debug:     engine.GetModule[debug.DebugModule](game, debug.ModuleID()),
		ecs:       engine.GetModule[ecs.ECSModule](game, ecs.ModuleID()),
	}
}

func (s *Scene) Enter() error {
	s.game.Renderer().SetClearColor(color.RGBA{R: 100, G: 149, B: 237, A: 255})

	assets := s.game.Assets()
	if err := assets.Load(content.AssetsFS(), gameplayManifest...); err != nil {
		return err
	}

	imageModule := engine.GetModule[images.ImageModule](s.game, images.ModuleID())
	imageAssets := imageModule.Assets()

	imgHandle, err := buildAnimations(assets, imageAssets, s.asprite.Library())
	if err != nil {
		return err
	}

	tiledModule := engine.GetModule[tiled.TiledModule](s.game, tiled.ModuleID())
	tiledAssets := tiledModule.Assets()

	tilemap, tilemapHandle, err := buildTilemap(s.ecs, tiledAssets, content.AssetsTiledGym04)
	if err != nil {
		return err
	}
	tilemapCenter := tilemap.Bounds().Center()

	s.tilemapEntry = tiledModule.CreateTilemap(s.ecs.World(),
		tiled.WithTilemapHandle(tilemapHandle),
	)
	s.cameraEntry = camera.NewCamera(s.ecs.World(),
		camera.AsMainCamera(),
		camera.WithZoom(0.25),
		camera.WithTransformOptions(
			transform.WithPosition(tilemapCenter.X, tilemapCenter.Y),
		),
	)
	s.spriteEntry = s.asprite.CreateSprite(s.ecs.World(),
		aseprite.WithImageOptions(
			images.WithHandle(imgHandle),
			images.WithAnchor(0.5, 1),
			images.WithTransformOptions(
				transform.WithPosition(tilemapCenter.X, tilemapCenter.Y),
			),
			images.WithRendererOptions(
				renderer.WithZIndex(1.5),
			),
		),
		aseprite.WithClip("Idle"),
		aseprite.Playing(),
	)

	width, height, _ := imageAssets.GetFrameSize(imgHandle)
	widthf, heightf := float64(width)*0.4, float64(height)*0.7

	collision.AddCollision(s.spriteEntry,
		collision.WithCollider(
			geom.Vec2{X: -widthf / 2, Y: -heightf},
			geom.Vec2{X: widthf / 2, Y: 0},
		),
	)

	movement.AddMovement(s.spriteEntry,
		movement.WithSpeed(player.GroundSpeed),
	)
	movement.AddGravity(s.spriteEntry)
	movement.AddJump(s.spriteEntry, player.JumpForce)

	s.ecs.Add(
		s.tilemapEntry,
		s.spriteEntry,
		s.cameraEntry,
	)
	return nil
}

func (s *Scene) Exit() error {

	return nil
}

func (s *Scene) Update(dt float64) (engine.SceneExitCode, error) {
	camera.RefreshCameraView(s.cameraEntry, s.game.Screen().SafeArea())

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return engine.SceneExitNone, ebiten.Termination
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	if inpututil.IsKeyJustPressed(ebiten.Key0) {
		s.debug.ToggleRendering()
	}
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		s.debug.ToggleTransformBounds()
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		s.debug.ToggleTransformInfo()
	}
	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		s.debug.ToggleCollisionBounds()
	}
	if inpututil.IsKeyJustPressed(ebiten.Key4) {
		s.debug.ToggleCollisionInfo()
	}
	if inpututil.IsKeyJustPressed(ebiten.Key5) {
		s.debug.ToggleCollisions()
	}
	if inpututil.IsKeyJustPressed(ebiten.Key6) {
		s.debug.ToggleStaticCollisionGrid()
	}
	if inpututil.IsKeyJustPressed(ebiten.Key7) {
		s.debug.ToggleDynamicCollisionGrid()
	}

	var moveX, moveY float64

	movement.ClearDirection(s.spriteEntry)
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		moveX -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		moveX += 1
	}
	movement.SetDirection(s.spriteEntry, moveX, moveY)

	jump := movement.GetJump(s.spriteEntry)
	gravity := movement.GetGravity(s.spriteEntry)

	if gravity.Enabled {
		if !jump.IsJumping && gravity.IsGrounded && inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			jump.IsJumping = true
			gravity.Velocity = -jump.Force
		}
	}

	return engine.SceneExitNone, nil
}

func (s *Scene) FixedUpdate(dt float64) error {
	world := s.ecs.World()

	movement.ApplyMovement(world, dt)
	s.ecs.Update(s.spriteEntry)

	s.collision.UpdateAllCollisions(s.spriteEntry)
	if hits, ok := s.collision.GetStaticCollisions(s.spriteEntry); ok {
		player.HandleStaticCollision(s.spriteEntry, hits)
	}
	return nil
}

func (s *Scene) LateUpdate(dt float64) error {
	s.ecs.ProcessPendingUpdates()

	player.UpdateAnimationState(s.spriteEntry, dt)

	viewport, _ := camera.GetView(s.cameraEntry)
	s.ecs.QueryAll(viewport, func(entry *donburi.Entry) {
		s.asprite.Systems().UpdateAnimation(entry, time.Duration(dt*float64(time.Second)))
	})
	return nil
}

func (s *Scene) Render(target *ebiten.Image) error {
	viewport, viewMatrix := camera.GetView(s.cameraEntry)
	s.debug.Render(target, viewport, viewMatrix)
	return nil
}

func buildAnimations(assets engine.Assets, imageAssets *images.ImageAssets, asepriteLibrary *aseprite.AsepriteLibrary) (uint64, error) {
	animationDataHandle, found := assets.GetDataHandle(content.AssetsAsepriteCaptainJson)
	if !found {
		return 0, engine.ErrAssetNotFound{Path: content.AssetsAsepriteCaptainJson.String()}
	}

	animationImageHandle, found := imageAssets.GetHandle(content.AssetsAsepriteCaptainImg)
	if !found {
		return 0, engine.ErrAssetNotFound{Path: content.AssetsAsepriteCaptainImg.String()}
	}

	animationData, _ := assets.GetData(animationDataHandle)
	asepriteLibrary.BuildAnimations(animationImageHandle, animationData)

	return animationImageHandle, nil
}

func buildTilemap(ecsModule ecs.ECSModule, tiledAssets *tiled.TiledAssets, tmxPath file.FilePath) (*tiled.Tilemap, uint64, error) {
	tmxHandle, found := tiledAssets.GetTmxHandle(tmxPath)
	if !found {
		return nil, 0, engine.ErrAssetNotFound{Path: tmxPath.String()}
	}

	tilemap, tmx := tiledAssets.BuildTilemap(tmxHandle)
	tmx.ObjectGroups.EachInGroup("collision", func(object *tiled.TmxObject) {
		entry := transform.NewTransform(ecsModule.World(),
			transform.WithPosition(object.X, object.Y),
			transform.WithBounds(
				geom.Vec2{},
				geom.Vec2{X: object.Width, Y: object.Height},
			),
		)
		collision.AddCollision(entry,
			collision.AsStatic(),
		)
		ecsModule.Add(entry)
	})

	return tilemap, tmxHandle, nil
}
