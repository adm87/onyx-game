package ecs

import (
	"github.com/adm87/onyx/pkg/engine/geom"
	"github.com/adm87/onyx/pkg/engine/partitioning/hashgrid"
	"github.com/yohamta/donburi"
)

type gridIndex struct {
	idx  uint64
	grid int
}

type ECSGrid struct {
	grid     []*hashgrid.HashGrid[donburi.Entity]
	indexing map[donburi.Entity]gridIndex

	queryGen  uint32
	querySeen map[donburi.Entity]uint32
}

func NewEntityGrid(resolutions ...int) *ECSGrid {
	grids := make([]*hashgrid.HashGrid[donburi.Entity], len(resolutions))
	for i, res := range resolutions {
		grids[i] = hashgrid.New[donburi.Entity](res)
	}
	return &ECSGrid{
		grid:      grids,
		indexing:  make(map[donburi.Entity]gridIndex),
		querySeen: make(map[donburi.Entity]uint32),
	}
}

func (m *ECSGrid) GetGrid(i int) *hashgrid.HashGrid[donburi.Entity] {
	if i < 0 || i >= len(m.grid) {
		return nil
	}
	return m.grid[i]
}

func (m *ECSGrid) Insert(entity donburi.Entity, area geom.AABB) uint64 {
	if index, exists := m.indexing[entity]; exists {
		return index.idx
	}
	grid, i := m.NearestGrid(area)

	id := grid.Insert(entity, area)
	m.indexing[entity] = gridIndex{
		idx:  id,
		grid: i,
	}

	return id
}

func (m *ECSGrid) Remove(entity donburi.Entity) {
	index, exists := m.indexing[entity]
	if !exists {
		return
	}

	grid := m.grid[index.grid]
	grid.Remove(index.idx)

	delete(m.indexing, entity)
}

func (m *ECSGrid) Update(entity donburi.Entity, area geom.AABB) uint64 {
	index, exists := m.indexing[entity]
	if !exists {
		return m.Insert(entity, area)
	}

	grid, i := m.NearestGrid(area)
	if i == index.grid {
		grid.Update(index.idx, area)
		return index.idx
	}

	oldGrid := m.grid[index.grid]
	oldGrid.Remove(index.idx)

	id := grid.Insert(entity, area)
	m.indexing[entity] = gridIndex{
		idx:  id,
		grid: i,
	}

	return id
}

func (m *ECSGrid) Query(area geom.AABB, callback func(donburi.Entity)) {
	m.queryGen++
	for _, grid := range m.grid {
		grid.Query(area, func(entity donburi.Entity) {
			if m.querySeen[entity] == m.queryGen {
				return
			}
			m.querySeen[entity] = m.queryGen
			callback(entity)
		})
	}
}

func (m *ECSGrid) NearestGrid(aabb geom.AABB) (*hashgrid.HashGrid[donburi.Entity], int) {
	resolution := int(max(aabb.Width(), aabb.Height()))
	for i, grid := range m.grid {
		if resolution <= grid.Resolution() {
			return grid, i
		}
	}
	i := len(m.grid) - 1
	return m.grid[i], i
}
