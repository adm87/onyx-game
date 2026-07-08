package player

import (
	"github.com/adm87/onyx/internal/game/movement"
	"github.com/adm87/onyx/pkg/modules/aseprite"
	"github.com/adm87/onyx/pkg/modules/ecs/transform"
	"github.com/yohamta/donburi"
)

func UpdateAnimationState(entry *donburi.Entry, dt float64) {
	gravity := movement.GetGravity(entry)
	move := movement.GetMovement(entry)

	if move.Direction.X < 0 {
		transform.SetScale(entry, -1, 1)
	} else if move.Direction.X > 0 {
		transform.SetScale(entry, 1, 1)
	}

	if gravity.Enabled && !gravity.IsGrounded {
		if gravity.Velocity > 30 {
			aseprite.SetClip(entry, "Fall")
			aseprite.SetLoops(entry, 1)
		} else {
			aseprite.SetClip(entry, "Jump")
			aseprite.SetLoops(entry, 1)
		}
	} else if move.Direction.X == 0 {
		aseprite.SetClip(entry, "Idle")
		aseprite.SetLoops(entry, -1)
	} else {
		aseprite.SetClip(entry, "Run")
		aseprite.SetLoops(entry, -1)
	}
}
