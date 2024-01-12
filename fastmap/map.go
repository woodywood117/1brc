package fastmap

import (
	"reflect"
	"strconv"
	"unsafe"
)

type Map[T any] struct {
	listHead *element[T]
	metadata *metadata[T]
	numItems uintptr
}

func New[T any]() *Map[T] {
	m := &Map[T]{listHead: newListHead[T]()}
	m.numItems = 0
	m.allocate(8)
	return m
}

func (m *Map[T]) Set(key string, value T) {
	var (
		h        = hash(key)
		valPtr   = &value
		alloc    *element[T]
		created  = false
		data     = m.metadata
		existing = data.indexElement(h)
	)

	if existing == nil || existing.keyHash > h {
		existing = m.listHead
	}
	if alloc, created = existing.inject(h, key, valPtr); alloc != nil {
		if created {
			m.numItems += 1
		}
	} else {
		for existing = m.listHead; alloc == nil; alloc, created = existing.inject(h, key, valPtr) {
		}
		if created {
			m.numItems += 1
		}
	}

	count := data.addItemToIndex(alloc)

	if (count*100)/uintptr(len(data.index)) > 50 {
		m.allocate(0) // double in size
	}
}

func (m *Map[T]) Get(key string) (value T, ok bool) {
	h := hash(key)
	// inline search
	for elem := m.metadata.indexElement(h); elem != nil && elem.keyHash <= h; elem = elem.nextPtr {
		if elem.key == key {
			return *elem.value, true
		}
	}
	ok = false
	return
}

func (m *Map[T]) allocate(newSize uintptr) {
	for {
		currentStore := m.metadata
		if newSize == 0 {
			newSize = uintptr(len(currentStore.index)) << 1
		} else {
			newSize = roundUpPower2(newSize)
		}

		index := make([]*element[T], newSize)
		header := (*reflect.SliceHeader)(unsafe.Pointer(&index))

		newdata := &metadata[T]{
			keyshifts: strconv.IntSize - log2(newSize),
			data:      unsafe.Pointer(header.Data),
			index:     index,
		}

		m.fillIndexItems(newdata) // re-index with longer and more widespread keys
		m.metadata = newdata

		if !((uintptr(m.numItems)*100)/newSize > 50) {
			return
		}
		newSize = 0 // 0 means double the current size
	}
}

func (m *Map[T]) fillIndexItems(mapData *metadata[T]) {
	var (
		first     = m.listHead.nextPtr
		item      = first
		lastIndex = uintptr(0)
	)
	for item != nil {
		index := item.keyHash >> mapData.keyshifts
		if item == first || index != lastIndex {
			mapData.addItemToIndex(item)
			lastIndex = index
		}
		item = item.nextPtr
	}
}

func roundUpPower2(i uintptr) uintptr {
	i--
	i |= i >> 1
	i |= i >> 2
	i |= i >> 4
	i |= i >> 8
	i |= i >> 16
	i |= i >> 32
	i++
	return i
}

func log2(i uintptr) (n uintptr) {
	for p := uintptr(1); p < i; p, n = p<<1, n+1 {
	}
	return
}
