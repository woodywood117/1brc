package fastmap

import (
	"strconv"
	"sync/atomic"
	"unsafe"
)

type metadata[T any] struct {
	keyshifts uintptr
	count     uintptr
	data      unsafe.Pointer // pointer to array of map indexes
	index     []*element[T]
}

const (
	intSizeBytes = strconv.IntSize >> 3
)

// TODO: Remove atomic operations
func (md *metadata[T]) addItemToIndex(item *element[T]) uintptr {
	index := item.keyHash >> md.keyshifts
	ptr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(md.data) + index*intSizeBytes))
	for {
		elem := (*element[T])(atomic.LoadPointer(ptr))
		if elem == nil {
			if atomic.CompareAndSwapPointer(ptr, nil, unsafe.Pointer(item)) {
				md.count += 1
				return md.count
			}
			continue
		}
		if item.keyHash < elem.keyHash {
			if !atomic.CompareAndSwapPointer(ptr, unsafe.Pointer(elem), unsafe.Pointer(item)) {
				continue
			}
		}
		return 0
	}
}

// TODO: Remove atomic operations
func (md *metadata[T]) indexElement(hashedKey uintptr) *element[T] {
	index := hashedKey >> md.keyshifts
	ptr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(md.data) + index*intSizeBytes))
	item := (*element[T])(atomic.LoadPointer(ptr))
	for (item == nil || hashedKey < item.keyHash) && index > 0 {
		index--
		ptr = (*unsafe.Pointer)(unsafe.Pointer(uintptr(md.data) + index*intSizeBytes))
		item = (*element[T])(atomic.LoadPointer(ptr))
	}
	return item
}
