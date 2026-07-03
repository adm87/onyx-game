package tiled

import "github.com/yohamta/donburi"

type TilemapOptions struct {
	Handle uint64
}

type TilemapOption func(*TilemapOptions)

var TilemapHandle = donburi.NewComponentType[uint64]()

func defaultTilemapOptions() *TilemapOptions {
	return &TilemapOptions{
		Handle: 0,
	}
}

func WithTilemapHandle(handle uint64) TilemapOption {
	return func(opts *TilemapOptions) {
		opts.Handle = handle
	}
}

func NewTilemap(world donburi.World, options ...TilemapOption) *donburi.Entry {
	return AddTilemap(world.Entry(world.Create(TilemapHandle)), options...)
}

func AddTilemap(entry *donburi.Entry, options ...TilemapOption) *donburi.Entry {
	opts := defaultTilemapOptions()
	for _, option := range options {
		option(opts)
	}

	donburi.Add(entry, TilemapHandle, &opts.Handle)

	return entry
}

func GetTilemapHandle(entry *donburi.Entry) uint64 {
	if !entry.HasComponent(TilemapHandle) {
		return 0
	}
	return *TilemapHandle.Get(entry)
}

func SetTilemapHandle(entry *donburi.Entry, handle uint64) {
	donburi.Add(entry, TilemapHandle, &handle)
}
