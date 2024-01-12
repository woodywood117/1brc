package cmd

import (
	"1brc/fastmap"
	"fmt"
	"testing"
)

func TestRun(t *testing.T) {
	run(nil, nil)
}

func TestMap(t *testing.T) {
	m := fastmap.New[int]()

	var total = 1000000

	for i := 0; i < total; i++ {
		m.Set(fmt.Sprintf("test-%d", i), i)
	}

	for i := 0; i < total+1; i++ {
		v, ok := m.Get(fmt.Sprintf("test-%d", i))
		if i == total {
			if ok {
				t.Errorf("Expected %v to not be in map", i)
			}
		} else {
			if !ok {
				t.Errorf("Expected %v to be in map", i)
			}
			if v != i {
				t.Errorf("Expected %v to be %v", i, v)
			}
		}
	}
}
