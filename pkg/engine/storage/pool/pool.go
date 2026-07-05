package pool

type Pool[T any] struct {
	free []T
	new  func() T
}

func NewPool[T any](newFunc func() T) *Pool[T] {
	return &Pool[T]{
		free: make([]T, 0),
		new:  newFunc,
	}
}

func (p *Pool[T]) Get() T {
	if len(p.free) > 0 {
		obj := p.free[len(p.free)-1]
		p.free = p.free[:len(p.free)-1]
		return obj
	}
	return p.new()
}

func (p *Pool[T]) Put(obj T) {
	p.free = append(p.free, obj)
}
