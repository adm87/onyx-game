package collision

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/events"
)

type CollisionEvent struct {
	Entry *donburi.Entry
	Hits  []*HitInfo
}

var (
	OnStaticCollisions  = events.NewEventType[*CollisionEvent]()
	OnDynamicCollisions = events.NewEventType[*CollisionEvent]()
)
