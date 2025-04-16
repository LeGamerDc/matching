package mwm

import (
	"container/heap"
)

// pairingHeaps

type pNode[T any] struct {
	child, next, prev int
	key               T
}

type pairingHeaps[T any] struct {
	heap []int
	node []pNode[T]

	less func(a, b T) bool
}

func newPairingHeaps[T any](h, n int, less func(T, T) bool) *pairingHeaps[T] {
	p := &pairingHeaps[T]{
		heap: make([]int, h),
		node: make([]pNode[T], n),
		less: less,
	}
	for i := range p.node {
		p.node[i].prev = -1
	}
	return p
}

func (p *pairingHeaps[T]) Clear(h int) {
	if p.heap[h] > 0 {
		p.clearRec(p.heap[h])
		p.heap[h] = 0
	}
}

func (p *pairingHeaps[T]) ClearAll() {
	for i := range p.heap {
		p.heap[i] = 0
	}
	for i := range p.node {
		var zero T
		p.node[i] = pNode[T]{prev: -1, next: 0, child: 0, key: zero}
	}
}

func (p *pairingHeaps[T]) Empty(h int) bool {
	return p.heap[h] == 0
}

func (p *pairingHeaps[T]) Used(v int) bool {
	return p.node[v].prev >= 0
}

func (p *pairingHeaps[T]) Min(h int) T {
	return p.node[p.heap[h]].key
}

func (p *pairingHeaps[T]) ArgMin(h int) int {
	return p.heap[h]
}

func (p *pairingHeaps[T]) Pop(h int) {
	p.Erase(h, p.heap[h])
}

func (p *pairingHeaps[T]) Push(h, v int, key T) {
	p.node[v] = pNode[T]{key: key}
	p.heap[h] = p.merge(p.heap[h], v)
}

func (p *pairingHeaps[T]) Erase(h, v int) {
	if !p.Used(v) {
		return
	}
	w := p.twoPassPairing(p.node[v].child)
	if p.node[v].prev == 0 {
		p.heap[h] = w
	} else {
		p.cut(v)
		p.heap[h] = p.merge(p.heap[h], w)
	}
	p.node[v].prev = -1
}

func (p *pairingHeaps[T]) DecreaseKey(h, v int, key T) {
	if !p.Used(v) {
		p.Push(h, v, key)
		return
	}
	if p.node[v].prev == 0 {
		p.node[v].key = key
	} else {
		p.cut(v)
		p.node[v].key = key
		p.heap[h] = p.merge(p.heap[h], v)
	}
}

func (p *pairingHeaps[T]) clearRec(v int) {
	for ; v > 0; v = p.node[v].next {
		if p.node[v].child > 0 {
			p.clearRec(p.node[v].child)
		}
		p.node[v].prev = -1
	}
}

func (p *pairingHeaps[T]) cut(v int) {
	n := &p.node[v]
	pv, nv := n.prev, n.next
	pn := &p.node[pv]
	if pn.child == v {
		pn.child = nv
	} else {
		pn.next = nv
	}
	p.node[nv].prev = pv
	n.next, n.prev = 0, 0
}

func (p *pairingHeaps[T]) merge(l, r int) int {
	if l == 0 {
		return r
	}
	if r == 0 {
		return l
	}
	// todo: 这里可能有bug 应该是 <=
	if p.less(p.node[r].key, p.node[l].key) {
		l, r = r, l
	}
	var lc int
	lc, p.node[r].next = p.node[l].child, p.node[l].child
	p.node[l].child, p.node[lc].prev = r, r
	p.node[r].prev = l
	return l
}

func (p *pairingHeaps[T]) twoPassPairing(root int) int {
	if root == 0 {
		return 0
	}
	var a, b, na, s, t int
	a, root = root, 0
	for a > 0 {
		b = p.node[a].next
		na = 0
		p.node[a].prev = 0
		p.node[a].next = 0
		if b > 0 {
			na = p.node[b].next
			p.node[b].prev = 0
			p.node[b].next = 0
		}
		a = p.merge(a, b)
		p.node[a].next = root
		root = a
		a = na
	}
	s, p.node[root].next = p.node[root].next, 0
	for s > 0 {
		t, p.node[s].next = p.node[s].next, 0
		root = p.merge(root, s)
		s = t
	}
	return root
}

// binaryHeap

type hNode[T any] struct {
	id    int
	value T
}

type binaryHeap[T any] struct {
	size  int
	node  []hNode[T]
	index []int

	less func(a, b T) bool
}

func newBinaryHeap[T any](n int, less func(T, T) bool) *binaryHeap[T] {
	return &binaryHeap[T]{
		size:  0,
		node:  make([]hNode[T], n+1),
		index: make([]int, n),
		less:  less,
	}
}

func (h *binaryHeap[T]) Size() int {
	return h.size
}

func (h *binaryHeap[T]) Empty() bool {
	return h.size == 0
}

func (h *binaryHeap[T]) Clear() {
	for h.size > 0 {
		h.index[h.node[h.size].id] = 0
		h.size--
	}
}

func (h *binaryHeap[T]) Min() T {
	return h.node[1].value
}

func (h *binaryHeap[T]) ArgMin() int {
	return h.node[1].id
}

func (h *binaryHeap[T]) GetV(id int) T {
	return h.node[h.index[id]].value
}

func (h *binaryHeap[T]) Pop() {
	if h.size > 0 {
		h.pop(1)
	}
}

func (h *binaryHeap[T]) Erase(id int) {
	if h.index[id] > 0 {
		h.pop(h.index[id])
	}
}

func (h *binaryHeap[T]) Has(id int) bool {
	return h.index[id] != 0
}

func (h *binaryHeap[T]) Update(id int, v T) {
	if !h.Has(id) {
		h.Push(id, v)
		return
	}
	up := h.less(v, h.node[h.index[id]].value)
	h.node[h.index[id]].value = v
	if up {
		h.upHeap(h.index[id])
	} else {
		h.downHeap(h.index[id])
	}
}

func (h *binaryHeap[T]) DecreaseKey(id int, v T) {
	if !h.Has(id) {
		h.Push(id, v)
		return
	}
	if h.less(v, h.node[h.index[id]].value) {
		h.node[h.index[id]].value = v
		h.upHeap(h.index[id])
	}
}

func (h *binaryHeap[T]) Push(id int, v T) {
	h.size++
	h.index[id] = h.size
	h.node[h.size] = hNode[T]{id: id, value: v}
	h.upHeap(h.size)
}

func (h *binaryHeap[T]) pop(pos int) {
	h.index[h.node[pos].id] = 0
	if pos == h.size {
		h.size--
		return
	}
	up := h.less(h.node[h.size].value, h.node[pos].value)
	h.node[pos] = h.node[h.size]
	h.size--
	h.index[h.node[pos].id] = pos
	if up {
		h.upHeap(pos)
	} else {
		h.downHeap(pos)
	}
}

func (h *binaryHeap[T]) downHeap(pos int) {
	k := pos
	nk := k
	for ; 2*k <= h.size; k = nk {
		if h.less(h.node[2*k].value, h.node[nk].value) {
			nk = 2 * k
		}
		if rk := 2*k + 1; rk <= h.size && h.less(h.node[rk].value, h.node[nk].value) {
			nk = rk
		}
		if nk == k {
			return
		}
		h.swap(k, nk)
	}
}

func (h *binaryHeap[T]) upHeap(pos int) {
	for k := pos; k > 1 && h.less(h.node[k].value, h.node[k>>1].value); k >>= 1 {
		h.swap(k, k>>1)
	}
}

func (h *binaryHeap[T]) swap(a, b int) {
	h.node[a], h.node[b] = h.node[b], h.node[a]
	h.index[h.node[a].id] = a
	h.index[h.node[b].id] = b
}

type fastHeap []edgeEvent

func (f *fastHeap) Less(i, j int) bool {
	return (*f)[i].time < (*f)[j].time
}

func (f *fastHeap) Swap(i, j int) {
	(*f)[i], (*f)[j] = (*f)[j], (*f)[i]
}

func (f *fastHeap) Push(x any) {
	*f = append(*f, x.(edgeEvent))
}

func (f *fastHeap) Pop() any {
	n := len(*f) - 1
	x := (*f)[n]
	*f = (*f)[:n]
	return x
}

func (f *fastHeap) Len() int {
	return len(*f)
}

var _ heap.Interface = (*fastHeap)(nil)
