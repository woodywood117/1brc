package fastmap

import (
	"strconv"
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

func (md *metadata[T]) addItemToIndex(item *element[T]) uintptr {
	index := item.keyHash >> md.keyshifts
	ptr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(md.data) + index*intSizeBytes))
	for {
		elem := (*element[T])(*ptr)
		if elem == nil {
			*ptr = unsafe.Pointer(item)
			md.count += 1
			return md.count
		}
		if item.keyHash < elem.keyHash {
			if *ptr == unsafe.Pointer(elem) {
				*ptr = unsafe.Pointer(item)
			} else {
				continue
			}
		}
		return 0
	}
}

func (md *metadata[T]) indexElement(hashedKey uintptr) *element[T] {
	index := hashedKey >> md.keyshifts
	ptr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(md.data) + index*intSizeBytes))
	item := (*element[T])(*ptr)
	for (item == nil || hashedKey < item.keyHash) && index > 0 {
		index--
		ptr = (*unsafe.Pointer)(unsafe.Pointer(uintptr(md.data) + index*intSizeBytes))
		item = (*element[T])(*ptr)
	}
	return item
}
