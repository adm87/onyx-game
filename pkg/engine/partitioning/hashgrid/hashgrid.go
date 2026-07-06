package hashgrid

import (
	"math"

	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/adm87/onyx/pkg/engine/storage/slotmap"
)

func encodeCell(cellX, cellY int) uint64 {
	return uint64(uint32(cellX))<<32 | uint64(uint32(cellY))
}

type HashGrid[T comparable] struct {
	store      *slotmap.SlotMap[T]
	queryGen   uint32
	resolution int
	cellCache  []uint64
	grid       map[uint64][]uint64
	cells      map[uint64][]uint64
	querySeen  map[uint64]uint32
}

func New[T comparable](resolution int) *HashGrid[T] {
	return &HashGrid[T]{
		store:      slotmap.New[T](0),
		cellCache:  make([]uint64, 0),
		grid:       make(map[uint64][]uint64),
		cells:      make(map[uint64][]uint64),
		querySeen:  make(map[uint64]uint32),
		resolution: resolution,
	}
}

func (sh *HashGrid[T]) Resolution() int {
	return sh.resolution
}

func (sh *HashGrid[T]) Insert(item T, area geom.AABB) uint64 {
	id := sh.store.Insert(item)

	sh.cellCache = sh.cacheCells(area, sh.cellCache[:0], false)
	for _, cell := range sh.cellCache {
		sh.grid[cell] = append(sh.grid[cell], id)
		sh.cells[id] = append(sh.cells[id], cell)
	}

	return id
}

func (sh *HashGrid[T]) Remove(id uint64) (T, bool) {
	item, exists := sh.store.Get(id)
	if !exists {
		var zero T
		return zero, false
	}

	cells, exists := sh.cells[id]
	if !exists {
		var zero T
		return zero, false
	}

	for _, cell := range cells {
		ids := sh.grid[cell]
		for i, cellID := range ids {
			if cellID == id {
				sh.grid[cell] = append(sh.grid[cell][:i], sh.grid[cell][i+1:]...)
				break
			}
		}
	}

	delete(sh.cells, id)
	sh.store.Delete(id)

	return item, true
}

func (sh *HashGrid[T]) Update(id uint64, area geom.AABB) {
	_, exists := sh.store.Get(id)
	if !exists {
		return
	}

	if cells, exists := sh.cells[id]; exists {
		for i := range cells {
			cell := cells[i]
			for j := range sh.grid[cell] {
				if sh.grid[cell][j] == id {
					sh.grid[cell] = append(sh.grid[cell][:j], sh.grid[cell][j+1:]...)
					break
				}
			}
			if len(sh.grid[cell]) == 0 {
				delete(sh.grid, cell)
			}
		}
	}

	sh.cellCache = sh.cacheCells(area, sh.cellCache[:0], false)
	sh.cells[id] = sh.cells[id][:0]

	for _, cell := range sh.cellCache {
		sh.grid[cell] = append(sh.grid[cell], id)
		sh.cells[id] = append(sh.cells[id], cell)
	}
}

func (sh *HashGrid[T]) Query(area geom.AABB, fn func(item T)) {
	sh.cellCache = sh.cacheCells(area, sh.cellCache[:0], true)
	sh.queryGen++
	for _, cell := range sh.cellCache {
		ids, exists := sh.grid[cell]
		if !exists {
			continue
		}

		for _, id := range ids {
			if sh.querySeen[id] == sh.queryGen {
				continue
			}
			sh.querySeen[id] = sh.queryGen

			item, exists := sh.store.Get(id)
			if !exists {
				continue
			}

			fn(item)
		}
	}
}

// GetCellRects returns the AABBs of the cells that intersect with the given area. The result slice is reused to avoid allocations.
func (sh *HashGrid[T]) GetCellRects(area geom.AABB, result []geom.AABB) []geom.AABB {
	sh.cellCache = sh.cacheCells(area, sh.cellCache[:0], false)
	for _, cell := range sh.cellCache {
		if _, exists := sh.grid[cell]; !exists {
			continue
		}

		cellX := int(int32(cell >> 32))
		cellY := int(int32(cell & 0xFFFFFFFF))

		result = append(result, geom.AABB{
			Min: geom.Vec2{
				X: float64(cellX * sh.resolution),
				Y: float64(cellY * sh.resolution),
			},
			Max: geom.Vec2{
				X: float64((cellX + 1) * sh.resolution),
				Y: float64((cellY + 1) * sh.resolution),
			},
		})
	}
	return result
}

func (sh *HashGrid[T]) cacheCells(area geom.AABB, cells []uint64, forQuery bool) []uint64 {
	var cellX1, cellY1, cellX2, cellY2 float64

	cellX1 = math.Floor(area.Min.X / float64(sh.resolution))
	cellY1 = math.Floor(area.Min.Y / float64(sh.resolution))

	if forQuery {
		cellX2 = math.Ceil(area.Max.X / float64(sh.resolution))
		cellY2 = math.Ceil(area.Max.Y / float64(sh.resolution))
	} else {
		cellX2 = math.Floor(area.Max.X / float64(sh.resolution))
		cellY2 = math.Floor(area.Max.Y / float64(sh.resolution))
	}

	minX := min(int(cellX1), int(cellX2))
	minY := min(int(cellY1), int(cellY2))
	maxX := max(int(cellX1), int(cellX2))
	maxY := max(int(cellY1), int(cellY2))

	for x := int(minX); x <= int(maxX); x++ {
		for y := int(minY); y <= int(maxY); y++ {
			cells = append(cells, encodeCell(x, y))
		}
	}
	return cells
}
