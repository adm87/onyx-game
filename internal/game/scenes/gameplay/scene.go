package gameplay

import (
	"image/color"
	"time"

	"github.com/adm87/onyx/content"
	"github.com/adm87/onyx/internal/game/components/movement"
	"github.com/adm87/onyx/pkg/engine"
	"github.com/adm87/onyx/pkg/engine/file"
	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/adm87/onyx/pkg/plugins/aseprite"
	"github.com/adm87/onyx/pkg/plugins/collision"
	"github.com/adm87/onyx/pkg/plugins/debug"
	"github.com/adm87/onyx/pkg/plugins/ecs"
	"github.com/adm87/onyx/pkg/plugins/ecs/camera"
	"github.com/adm87/onyx/pkg/plugins/ecs/renderer"
	"github.com/adm87/onyx/pkg/plugins/ecs/transform"
	"github.com/adm87/onyx/pkg/plugins/images"
	"github.com/adm87/onyx/pkg/plugins/tiled"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/donburi"
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

	asepritePlugin  aseprite.AsepritePlugin
	collisionPlugin collision.CollisionPlugin
	debugPlugin     debug.DebugPlugin
	ecsPlugin       ecs.ECSPlugin

	debugToggleRendering bool
}

func NewScene(game engine.Game) *Scene {
	return &Scene{
		game:            game,
		asepritePlugin:  engine.GetPlugin[aseprite.AsepritePlugin](game, aseprite.PluginID()),
		collisionPlugin: engine.GetPlugin[collision.CollisionPlugin](game, collision.PluginID()),
		debugPlugin:     engine.GetPlugin[debug.DebugPlugin](game, debug.PluginID()),
		ecsPlugin:       engine.GetPlugin[ecs.ECSPlugin](game, ecs.PluginID()),
	}
}

func (s *Scene) Enter() error {
	s.game.Renderer().SetBackgroundColor(color.RGBA{R: 100, G: 149, B: 237, A: 255})

	assets := s.game.Assets()
	if err := assets.Load(content.AssetsFS(), gameplayManifest...); err != nil {
		return err
	}

	imagePlugin := engine.GetPlugin[images.ImagePlugin](s.game, images.PluginID())
	imageAssets := imagePlugin.Assets()

	imgHandle, err := buildAnimations(assets, imageAssets, s.asepritePlugin.Library())
	if err != nil {
		return err
	}

	tiledPlugin := engine.GetPlugin[tiled.TiledPlugin](s.game, tiled.PluginID())
	tiledAssets := tiledPlugin.Assets()

	tilemap, tilemapHandle, err := buildTilemap(s.ecsPlugin, tiledAssets, content.AssetsTiledGym04)
	if err != nil {
		return err
	}
	tilemapCenter := tilemap.Bounds().Center()

	s.tilemapEntry = tiledPlugin.CreateTilemap(s.ecsPlugin.World(),
		tiled.WithTilemapHandle(tilemapHandle),
	)

	s.cameraEntry = transform.NewTransform(s.ecsPlugin.World())
	s.cameraEntry.AddComponent(camera.MainCamera)

	transform.SetPosition(s.cameraEntry, tilemapCenter.X, tilemapCenter.Y)
	camera.SetZoom(s.cameraEntry, 0.25)

	s.spriteEntry = s.asepritePlugin.CreateSprite(s.ecsPlugin.World(),
		aseprite.WithImageOptions(
			images.WithHandle(imgHandle),
			images.WithAnchor(0.5, 1),
		),
		aseprite.WithClip("Idle"),
		aseprite.Playing(),
	)

	renderer.SetZIndex(s.spriteEntry, 1.5)
	transform.SetPosition(s.spriteEntry, tilemapCenter.X, tilemapCenter.Y)

	collision.AddCollision(s.spriteEntry)

	movement.AddMovement(s.spriteEntry,
		movement.WithSpeed(100),
	)

	s.ecsPlugin.Add(
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
		s.debugPlugin.ToggleRendering()
	}
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		s.debugPlugin.ToggleTransformBounds()
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		s.debugPlugin.ToggleTransformInfo()
	}
	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		s.debugPlugin.ToggleCollisionBounds()
	}
	if inpututil.IsKeyJustPressed(ebiten.Key4) {
		s.debugPlugin.ToggleCollisionInfo()
	}

	var moveX, moveY float64

	movement.ClearDirection(s.spriteEntry)
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		moveY -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		moveY += 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		moveX -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		moveX += 1
	}
	movement.SetDirection(s.spriteEntry, moveX, moveY)

	return engine.SceneExitNone, nil
}

func (s *Scene) FixedUpdate(dt float64) error {
	movement.ApplyMovement(s.ecsPlugin.World(), dt)
	if movement.IsMoving(s.spriteEntry) {
		s.ecsPlugin.Update(s.spriteEntry)
	}
	return nil
}

func (s *Scene) LateUpdate(dt float64) error {
	direction := movement.GetDirection(s.spriteEntry)
	if direction.X == 0 {
		aseprite.SetClip(s.spriteEntry, "Idle")
	} else {
		if direction.X < 0 {
			transform.SetScale(s.spriteEntry, -1, 1)
		} else if direction.X > 0 {
			transform.SetScale(s.spriteEntry, 1, 1)
		}
		aseprite.SetClip(s.spriteEntry, "Run")
	}

	spriteX, spriteY := transform.GetPosition(s.spriteEntry)
	transform.SetPosition(s.cameraEntry, spriteX, spriteY)

	viewport, _ := camera.GetView(s.cameraEntry)

	asepriteSystems := s.asepritePlugin.Systems()
	s.ecsPlugin.QueryAll(viewport, func(entry *donburi.Entry) {
		asepriteSystems.UpdateAnimation(entry, time.Duration(dt*float64(time.Second)))
	})
	return nil
}

func (s *Scene) Render(target *ebiten.Image) error {
	viewport, viewMatrix := camera.GetView(s.cameraEntry)
	s.debugPlugin.Render(target, viewport, viewMatrix)
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

func buildTilemap(ecsPlugin ecs.ECSPlugin, tiledAssets *tiled.TiledAssets, tmxPath file.FilePath) (*tiled.Tilemap, uint64, error) {
	tmxHandle, found := tiledAssets.GetTmxHandle(tmxPath)
	if !found {
		return nil, 0, engine.ErrAssetNotFound{Path: tmxPath.String()}
	}

	tilemap, tmx := tiledAssets.BuildTilemap(tmxHandle)
	tmx.ObjectGroups.EachInGroup("collision", func(object *tiled.TmxObject) {
		entry := transform.NewTransform(ecsPlugin.World(),
			transform.WithPosition(object.X, object.Y),
			transform.WithBounds(
				geom.Vec2{},
				geom.Vec2{X: object.Width, Y: object.Height},
			),
		)
		collision.AddCollision(entry,
			collision.AsStatic(),
		)
		ecsPlugin.Add(entry)
	})

	return tilemap, tmxHandle, nil
}
