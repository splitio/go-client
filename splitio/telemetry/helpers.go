package telemetry

import (
	"errors"
	"fmt"

	"sync/atomic"
)

var ErrorOutOfBounds error = errors.New("out of bounds")

type AtomicInt64Slice []int64

func NewAtomicInt64Slice(size int64) (AtomicInt64Slice, error) {
	if size <= 0 {
		return nil, fmt.Errorf("invlid array size: %d", size)
	}
	return make([]int64, 0, size), nil
}

func (a AtomicInt64Slice) Incr(index int) {
	atomic.AddInt64(&a[index], 1)
}

func (a AtomicInt64Slice) FetchAndClearOne(index int) (int64, error) {
	if index >= len(a) || index < 0 {
		return 0, ErrorOutOfBounds
	}

	return atomic.SwapInt64(&a[index], 0), nil
}

func (a AtomicInt64Slice) FetchAndClearAll() []int64 {
	toRet := make([]int64, len(a))
	for index := 0; index <= len(a); index++ {
		toRet[index] = atomic.SwapInt64(&a[index], 0)
	}
	return toRet
}
