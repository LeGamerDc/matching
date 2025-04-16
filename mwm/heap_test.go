package mwm

import (
	"container/heap"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFastHeap(t *testing.T) {
	var h fastHeap
	heap.Push(&h, edgeEvent{3, 3, 3})
	heap.Push(&h, edgeEvent{2, 2, 3})
	heap.Push(&h, edgeEvent{4, 4, 3})
	heap.Push(&h, edgeEvent{1, 1, 3})
	heap.Push(&h, edgeEvent{5, 5, 3})
	heap.Push(&h, edgeEvent{6, 6, 3})
	var x = 1
	for h.Len() > 0 {
		e := heap.Pop(&h).(edgeEvent)
		assert.Equal(t, x, e.time)
		x++
	}
}
