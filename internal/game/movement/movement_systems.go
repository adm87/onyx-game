package movement

import (
	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/adm87/onyx/pkg/modules/ecs/transform"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"
)

var (
	movementQuery = donburi.NewQuery(
		filter.Contains(
			Movement,
			transform.Transform,
		),
	)
)

func ApplyMovement(ecs donburi.World, dt float64) {
	movementQuery.Each(ecs, func(entry *donburi.Entry) {
		var vel geom.Vec2

		movement := GetMovement(entry)
		vel.X += movement.Direction.X * movement.Speed

		gravity := GetGravity(entry)
		if gravity.Enabled {
			gravity.IsGrounded = false

			gravity.Velocity += GravityAcceleration * dt
			if gravity.Velocity > TerminalVelocity {
				gravity.Velocity = TerminalVelocity
			}

			vel.Y = gravity.Velocity
		} else {
			vel.Y += movement.Direction.Y * movement.Speed
		}

		transform.Translate(entry, vel.X*dt, vel.Y*dt)
	})
}
