package movement

import (
	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/yohamta/donburi"
)

type MovementOptions struct {
	Direction geom.Vec2
	Speed     float64
}

type MovementOption func(*MovementOptions)

type MovementModel struct {
	Direction geom.Vec2
	Speed     float64
}

var (
	Movement = donburi.NewComponentType[MovementModel]()
)

func WithDirection(direction geom.Vec2) MovementOption {
	return func(o *MovementOptions) {
		o.Direction = direction
	}
}

func WithSpeed(speed float64) MovementOption {
	return func(o *MovementOptions) {
		o.Speed = speed
	}
}

func AddMovement(entry *donburi.Entry, opts ...MovementOption) {
	options := &MovementOptions{
		Direction: geom.Vec2{X: 0, Y: 0},
		Speed:     0,
	}
	for _, opt := range opts {
		opt(options)
	}
	SetMovement(entry, &MovementModel{
		Direction: options.Direction,
		Speed:     options.Speed,
	})
}

func GetMovement(entry *donburi.Entry) *MovementModel {
	if !entry.HasComponent(Movement) {
		return &MovementModel{
			Direction: geom.Vec2{X: 0, Y: 0},
			Speed:     0,
		}
	}
	return Movement.Get(entry)
}

func SetMovement(entry *donburi.Entry, movement *MovementModel) {
	donburi.Add(entry, Movement, movement)
}

func GetDirection(entry *donburi.Entry) geom.Vec2 {
	if !entry.HasComponent(Movement) {
		return geom.Vec2{X: 0, Y: 0}
	}
	return Movement.Get(entry).Direction
}

func SetDirection(entry *donburi.Entry, x, y float64) {
	if !entry.HasComponent(Movement) {
		return
	}
	movement := Movement.Get(entry)
	movement.Direction = geom.Vec2{X: x, Y: y}
}

func ClearDirection(entry *donburi.Entry) {
	if !entry.HasComponent(Movement) {
		return
	}
	movement := Movement.Get(entry)
	movement.Direction = geom.Vec2{X: 0, Y: 0}
}

func GetSpeed(entry *donburi.Entry) float64 {
	if !entry.HasComponent(Movement) {
		return 0
	}
	return Movement.Get(entry).Speed
}

func SetSpeed(entry *donburi.Entry, speed float64) {
	if !entry.HasComponent(Movement) {
		return
	}
	movement := Movement.Get(entry)
	movement.Speed = speed
}

func ClearSpeed(entry *donburi.Entry) {
	if !entry.HasComponent(Movement) {
		return
	}
	movement := Movement.Get(entry)
	movement.Speed = 0
}

func ClearMovement(entry *donburi.Entry) {
	if !entry.HasComponent(Movement) {
		return
	}
	movement := Movement.Get(entry)
	movement.Direction = geom.Vec2{X: 0, Y: 0}
	movement.Speed = 0
}

func IsMoving(entry *donburi.Entry) bool {
	if !entry.HasComponent(Movement) {
		return false
	}
	movement := Movement.Get(entry)
	return movement.Speed > 0 && (movement.Direction.X != 0 || movement.Direction.Y != 0)
}
