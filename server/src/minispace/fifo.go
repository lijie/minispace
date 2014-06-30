package minispace

import "sync"

type Fifo struct {
	// need lock ?
	lock sync.Mutex
	// need atomic int ?
	pt int
	gt int
	size int
	value []interface{}
}

func (f *Fifo) Put(v interface{}) error {
	if ((f.pt + 1) & (f.size - 1)) == f.gt {
		return ErrFifoIsFull
	}

	f.value[f.pt] = v
	f.pt = (f.pt + 1) & (f.size - 1)
	return nil
}

func (f *Fifo) Get() (interface{}, error) {
	if f.pt == f.gt {
		return nil, ErrFifoIsEmpty
	}

	v := f.value[f.gt]
	f.gt = (f.gt + 1) & (f.size - 1)
	return v, nil
}

func NewFifo(size int) (*Fifo, error) {
	if size & (size - 1) != 0 {
		return nil, ErrFifoInvalidSize
	}

	f := &Fifo{
		size: size,
		value: make([]interface{}, size),
	}

	return f, nil
}
