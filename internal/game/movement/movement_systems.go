package movement

import (
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
		movement := GetMovement(entry)
		if movement.Speed > 0 && (movement.Direction.X != 0 || movement.Direction.Y != 0) {
			move := movement.Direction.Normalize().Mul(movement.Speed * dt)
			transform.Translate(entry, move.X, move.Y)
		}
	})
}
