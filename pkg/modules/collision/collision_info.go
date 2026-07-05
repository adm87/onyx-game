package collision

import (
	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/yohamta/donburi"
)

type HitInfo struct {
	Other *donburi.Entry

	ColliderA geom.AABB
	ColliderB geom.AABB
	Overlap   geom.AABB

	Normal geom.Vec2
	Depth  float64
}

type CollisionInfo struct {
	info []*HitInfo
}
