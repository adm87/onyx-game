package player

import (
	"github.com/adm87/onyx/internal/game/movement"
	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/adm87/onyx/pkg/modules/collision"
	"github.com/adm87/onyx/pkg/modules/ecs/transform"
	"github.com/yohamta/donburi"
)

func HandleStaticCollision(entry *donburi.Entry, hits []*collision.HitInfo) {
	if len(hits) == 0 {
		return
	}

	horizontal, vertical := getDeepestCollisions(hits)
	correction := geom.Vec2{}

	gravity := movement.GetGravity(entry)
	jump := movement.GetJump(entry)

	if horizontal != nil {
		correction.X = horizontal.Normal.X * horizontal.Depth
	}
	if vertical != nil {
		corY := vertical.Normal.Y * vertical.Depth

		if gravity.Enabled {
			if vertical.Normal.Y < 0 && gravity.Velocity > 0 {
				jump.IsJumping = false
				gravity.IsGrounded = true
				gravity.Velocity = 0
				correction.Y = corY
			} else if vertical.Normal.Y > 0 && gravity.Velocity < 0 {
				gravity.Velocity = 0
				correction.Y = corY
			}
		} else {
			correction.Y = corY
		}
	}

	transform.Translate(entry, correction.X, correction.Y)
}

func getDeepestCollisions(infos []*collision.HitInfo) (horizontal, vertical *collision.HitInfo) {
	for _, info := range infos {
		if info.Normal.X != 0 && (horizontal == nil || info.Depth > horizontal.Depth) {
			horizontal = info
		} else if info.Normal.Y != 0 && (vertical == nil || info.Depth > vertical.Depth) {
			vertical = info
		}
	}
	return horizontal, vertical
}
