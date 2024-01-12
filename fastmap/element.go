package fastmap

type element[T any] struct {
	keyHash uintptr
	key     string
	nextPtr *element[T]
	value   *T
}

func newListHead[T any]() *element[T] {
	e := &element[T]{keyHash: 0, key: *new(string)}
	e.nextPtr = nil
	e.value = new(T)
	return e
}

func (self *element[T]) addBefore(allocatedElement, before *element[T]) bool {
	if self.nextPtr != before {
		return false
	}
	allocatedElement.nextPtr = before
	if self.nextPtr == before {
		self.nextPtr = allocatedElement
		return true
	}
	return false
}

func (self *element[T]) inject(c uintptr, key string, value *T) (*element[T], bool) {
	var (
		alloc             *element[T]
		left, curr, right = self.search(c, key)
	)
	if curr != nil {
		curr.value = value
		return curr, false
	}
	if left != nil {
		alloc = &element[T]{keyHash: c, key: key}
		alloc.value = value
		if left.addBefore(alloc, right) {
			return alloc, true
		}
	}
	return nil, false
}

func (self *element[T]) search(c uintptr, key string) (*element[T], *element[T], *element[T]) {
	var (
		left, right *element[T]
		curr        = self
	)
	for {
		if curr == nil {
			return left, curr, right
		}
		right = curr.nextPtr
		if c < curr.keyHash {
			right = curr
			curr = nil
			return left, curr, right
		} else if c == curr.keyHash && key == curr.key {
			return left, curr, right
		}
		left = curr
		curr = left.nextPtr
		right = nil
	}
}
