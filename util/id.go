package util

import "sync"

type Id struct {
	sync.Mutex
	id uint32
}

func NewId(defaultId uint32) *Id {
	o := &Id{id: defaultId}

	return o
}

func (i *Id) IncrementAndGet() uint32 {
	i.Lock()
	defer i.Unlock()

	i.id++
	return i.id
}

func (i *Id) DecrementAndGet() uint32 {

	i.Lock()
	defer i.Unlock()

	i.id--
	return i.id
}
