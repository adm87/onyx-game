package camera

import (
	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/adm87/onyx/pkg/modules/ecs/transform"
	"github.com/yohamta/donburi"
)

func RefreshCameraView(entry *donburi.Entry, screenBounds geom.AABB) {
	center := screenBounds.Center()
	transform.SetOrigin(entry, center.X, center.Y)
	transform.SetBounds(entry, &screenBounds)
}
