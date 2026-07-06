package movement

import (
	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/yohamta/donburi"
)

const (
	GravityAcceleration = 98.1 * 2
	TerminalVelocity    = 150.0
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

type GravityModel struct {
	Velocity   float64
	IsGrounded bool
	Enabled    bool
}

type JumpModel struct {
	Force     float64
	IsJumping bool
}

var (
	Movement = donburi.NewComponentType[MovementModel]()
	Gravity  = donburi.NewComponentType[GravityModel]()
	Jump     = donburi.NewComponentType[JumpModel]()
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

func AddGravity(entry *donburi.Entry) {
	SetGravity(entry, &GravityModel{
		IsGrounded: false,
		Enabled:    true,
	})
}

func AddJump(entry *donburi.Entry, force float64) {
	SetJump(entry, &JumpModel{
		Force:     force,
		IsJumping: false,
	})
}

func GetJump(entry *donburi.Entry) *JumpModel {
	if !entry.HasComponent(Jump) {
		return &JumpModel{
			Force:     0,
			IsJumping: false,
		}
	}
	return Jump.Get(entry)
}

func SetJump(entry *donburi.Entry, jump *JumpModel) {
	donburi.Add(entry, Jump, jump)
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

	move := Movement.Get(entry)
	grav := GetGravity(entry)

	if grav.Enabled {
		return move.Direction.X != 0 || !grav.IsGrounded
	}
	return move.Direction.X != 0 || move.Direction.Y != 0
}

func GetGravity(entry *donburi.Entry) *GravityModel {
	if !entry.HasComponent(Gravity) {
		return &GravityModel{
			Velocity:   0,
			IsGrounded: false,
			Enabled:    false,
		}
	}
	return Gravity.Get(entry)
}

func SetGravity(entry *donburi.Entry, gravity *GravityModel) {
	donburi.Add(entry, Gravity, gravity)
}

func IsGrounded(entry *donburi.Entry) bool {
	if !entry.HasComponent(Gravity) {
		return false
	}
	gravity := Gravity.Get(entry)
	return gravity.IsGrounded
}

func SetGrounded(entry *donburi.Entry, grounded bool) {
	if !entry.HasComponent(Gravity) {
		return
	}
	gravity := Gravity.Get(entry)
	gravity.IsGrounded = grounded
}

func GetGravityEnabled(entry *donburi.Entry) bool {
	if !entry.HasComponent(Gravity) {
		return false
	}
	return Gravity.Get(entry).Enabled
}

func SetGravityEnabled(entry *donburi.Entry, enabled bool) {
	if !entry.HasComponent(Gravity) {
		return
	}
	gravity := Gravity.Get(entry)
	gravity.Enabled = enabled
}
