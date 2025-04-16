package mwm

import (
	"container/heap"
)

type label int

const (
	kSeparated label = iota - 2
	kInner
	kFree
	kOuter
)

const inf int = 1 << 30

type edgeEvent struct {
	time     int
	from, to int
}

func (e edgeEvent) Less(b edgeEvent) bool {
	return e.time < b.time
}

type event struct {
	time int
	id   int
}

func (e event) Less(b event) bool {
	return e.time < b.time
}

type edge struct {
	from, to int
	cost     int
}

type link struct {
	from, to int
}

type nodeLink struct {
	b, v int
}

type node struct {
	parent, size int
	link         [2]nodeLink
}

func newNode(u int) node {
	return node{
		parent: 0, size: 1,
		link: [2]nodeLink{{u, u}, {u, u}},
	}
}

func (n node) nextV() int {
	return n.link[0].v
}

func (n node) nextB() int {
	return n.link[0].b
}

func (n node) prevV() int {
	return n.link[1].v
}

func (n node) prevB() int {
	return n.link[1].b
}

type B5 struct {
	n, b, s int
	ofs     []int
	edges   []edge
	input   []edge

	que                 []int
	mate, surface, base []int
	link                []link
	label               []label
	potential           []int

	unusedBid    []int
	unusedBidIdx int
	node         []node

	heavy, group             []int
	timeCreated, lazy, slack []int
	bestFrom                 []int

	timeCurrent int
	event1      event
	heap2       *binaryHeap[edgeEvent]
	heap2s      *pairingHeaps[edgeEvent]
	heap3       fastHeap
	heap4       *binaryHeap[int]
}

func New(n int) *B5 {
	b := (n - 1) / 2
	s := n + b + 1
	b5 := &B5{
		n:            n,
		b:            b,
		s:            s,
		ofs:          make([]int, n+2),
		mate:         make([]int, s),
		surface:      make([]int, s),
		base:         make([]int, s),
		link:         make([]link, s),
		label:        make([]label, s),
		potential:    make([]int, s),
		unusedBid:    make([]int, b),
		unusedBidIdx: b,
		node:         make([]node, s),
		heavy:        make([]int, s),
		group:        make([]int, s),
		timeCreated:  make([]int, s),
		lazy:         make([]int, s),
		slack:        make([]int, s),
		bestFrom:     make([]int, s),
		timeCurrent:  0,
		event1:       event{time: inf, id: 0},
		heap2: newBinaryHeap[edgeEvent](s, func(x edgeEvent, y edgeEvent) bool {
			return x.time < y.time
		}),
		heap2s: newPairingHeaps[edgeEvent](s, s, func(x edgeEvent, y edgeEvent) bool {
			return x.time < y.time
		}),
		heap3: make(fastHeap, 0, 2000),
		heap4: newBinaryHeap[int](s, func(x int, y int) bool {
			return x < y
		}),
	}
	for i := 0; i < s; i++ {
		b5.label[i] = kFree
		b5.base[i] = i
		b5.surface[i] = i
		b5.slack[i] = inf
		b5.group[i] = i
		if i != 0 {
			b5.node[i] = newNode(i)
		}
	}
	for i := 0; i < b; i++ {
		b5.unusedBid[i] = n + b - i
	}
	b5.resetTime()
	return b5
}

func (m *B5) AddEdge(u, v, w int) {
	m.input = append(m.input, edge{u, v, w})
}

func (m *B5) initialize() {
	m.edges = make([]edge, 2*len(m.input))
	for _, e := range m.input {
		m.ofs[e.from+1]++
		m.ofs[e.to+1]++
	}
	for i := 1; i <= m.n+1; i++ {
		m.ofs[i] += m.ofs[i-1]
	}
	for _, e := range m.input {
		m.edges[m.ofs[e.from]] = edge{from: e.from, to: e.to, cost: e.cost * 2}
		m.ofs[e.from]++
		m.edges[m.ofs[e.to]] = edge{from: e.to, to: e.from, cost: e.cost * 2}
		m.ofs[e.to]++
	}
	for i := m.n + 1; i > 0; i-- {
		m.ofs[i] = m.ofs[i-1]
	}
	m.ofs[0] = 0
}

func (m *B5) push(x int) {
	m.que = append(m.que, x)
}

func (m *B5) pop() (x int) {
	x, m.que = m.que[0], m.que[1:]
	return
}

func (m *B5) setPotential() {
	for u := 1; u <= m.n; u++ {
		maxC := 0
		for eid := m.ofs[u]; eid < m.ofs[u+1]; eid++ {
			maxC = max(maxC, m.edges[eid].cost)
		}
		m.potential[u] = maxC >> 1
	}
}

func (m *B5) findMaximumMatching() {
	var e *edge
	var v int
	for u := 1; u <= m.n; u++ {
		if m.mate[u] == 0 {
			for eid := m.ofs[u]; eid < m.ofs[u+1]; eid++ {
				e = &m.edges[eid]
				v = e.to
				if m.mate[v] > 0 || m.reducedCost(u, v, e) > 0 {
					continue
				}
				m.mate[u] = v
				m.mate[v] = u
				break
			}
		}
	}
}

func (m *B5) computeOptimalValue() int {
	ret := 0
	for u := 1; u <= m.n; u++ {
		if m.mate[u] > u {
			mc := 0
			for eid := m.ofs[u]; eid < m.ofs[u+1]; eid++ {
				if m.edges[eid].to == m.mate[u] {
					mc = max(mc, m.edges[eid].cost)
				}
			}
			ret += mc
		}
	}
	return ret >> 1
}

func (m *B5) reducedCost(u, v int, e *edge) int {
	return m.potential[u] + m.potential[v] - e.cost
}

func (m *B5) rematch(v, w int) {
	t := m.mate[v]
	m.mate[v] = w
	if m.mate[t] != v {
		return
	}
	if m.link[v].to == m.surface[m.link[v].to] {
		m.mate[t] = m.link[v].from
		m.rematch(m.mate[t], t)
	} else {
		x := m.link[v].from
		y := m.link[v].to
		m.rematch(x, y)
		m.rematch(y, x)
	}
}

func (m *B5) fixMateAndBase(b int) {
	if b <= m.n {
		return
	}
	bv := m.base[b]
	mv := m.node[bv].link[0].v
	bmv := m.node[bv].link[0].b
	d := or(m.node[bmv].link[1].v == m.mate[mv], 0, 1)
	for {
		mv = m.node[bv].link[d].v
		bmv = m.node[bv].link[d].b
		if m.node[bmv].link[1-d].v != m.mate[mv] {
			break
		}
		m.fixMateAndBase(bv)
		m.fixMateAndBase(bmv)
		bv = m.node[bmv].link[d].b
	}
	m.base[b] = bv
	m.fixMateAndBase(bv)
	m.mate[b] = m.mate[bv]
}

func (m *B5) resetTime() {
	m.timeCurrent = 0
	m.event1 = event{time: inf, id: 0}
}

func (m *B5) resetBlossom(b int) {
	m.label[b] = kFree
	m.link[b].from = 0
	m.slack[b] = inf
	m.lazy[b] = 0
}

func (m *B5) resetAll() {
	m.label[0] = kFree
	m.link[0].from = 0
	for v := 1; v <= m.n; v++ {
		if m.label[v] == kOuter {
			m.potential[v] -= m.timeCurrent
		} else {
			bv := m.surface[v]
			m.potential[v] += m.lazy[bv]
			if m.label[bv] == kInner {
				m.potential[v] += m.timeCurrent - m.timeCreated[bv]
			}
		}
		m.resetBlossom(v)
	}
	b := m.n + 1
	r := m.b - m.unusedBidIdx
	for ; r > 0 && b < m.s; b++ {
		if m.base[b] != b {
			if m.surface[b] == b {
				m.fixMateAndBase(b)
				if m.label[b] == kOuter {
					m.potential[b] += (m.timeCurrent - m.timeCreated[b]) << 1
				} else if m.label[b] == kInner {
					m.fixBlossomPotential(b, kInner)
				} else {
					m.fixBlossomPotential(b, kFree)
				}
			}
			m.heap2s.Clear(b)
			m.resetBlossom(b)
			r--
		}
	}
	m.que = m.que[:0]
	m.resetTime()
	m.heap2.Clear()
	m.heap3 = m.heap3[:0]
	m.heap4.Clear()
}

func (m *B5) doEdmondsSearch(root int) {
	if m.potential[root] == 0 {
		return
	}
	m.linkBlossom(m.surface[root], link{from: 0, to: 0})
	m.pushOuterAndFixPotentials(m.surface[root], 0)
	for augmented := false; !augmented; {
		if augmented = m.augment(root); augmented {
			break
		}
		augmented = m.adjustDualVariables(root)
	}
	m.resetAll()
}

func (m *B5) fixBlossomPotential(b int, lab label) int {
	d := m.lazy[b]
	m.lazy[b] = 0
	if lab == kInner {
		dt := m.timeCurrent - m.timeCreated[b]
		if b > m.n {
			m.potential[b] -= dt << 1
		}
		d += dt
	}
	return d
}

func (m *B5) updateHeap2(x, y, by, t int, lab label) {
	if t >= m.slack[y] {
		return
	}
	m.slack[y] = t
	m.bestFrom[y] = x
	if y == by {
		if lab != kInner {
			m.heap2.DecreaseKey(y, edgeEvent{time: t + m.lazy[y], from: x, to: y})
		}
	} else {
		gy := m.group[y]
		if gy != y {
			if t >= m.slack[gy] {
				return
			}
			m.slack[gy] = t
		}
		m.heap2s.DecreaseKey(by, gy, edgeEvent{time: t, from: x, to: y})
		if lab == kInner {
			return
		}
		mi := m.heap2s.Min(by)
		m.heap2.DecreaseKey(by, edgeEvent{time: mi.time + m.lazy[by], from: mi.from, to: mi.to})
	}
}

func (m *B5) activateHeap2Node(b int) {
	if b <= m.n {
		if m.slack[b] < inf {
			m.heap2.Push(b, edgeEvent{time: m.slack[b] + m.lazy[b], from: m.bestFrom[b], to: b})
		}
	} else {
		if m.heap2s.Empty(b) {
			return
		}
		mi := m.heap2s.Min(b)
		m.heap2.Push(b, edgeEvent{time: mi.time + m.lazy[b], from: mi.from, to: mi.to})
	}
}

func (m *B5) swapBlossom(a, b int) {
	m.base[a], m.base[b] = m.base[b], m.base[a]
	if m.base[a] == a {
		m.base[a] = b
	}
	m.heavy[a], m.heavy[b] = m.heavy[b], m.heavy[a]
	if m.heavy[a] == a {
		m.heavy[a] = b
	}
	m.link[a], m.link[b] = m.link[b], m.link[a]
	m.mate[a], m.mate[b] = m.mate[b], m.mate[a]
	m.potential[a], m.potential[b] = m.potential[b], m.potential[a]
	m.lazy[a], m.lazy[b] = m.lazy[b], m.lazy[a]
	m.timeCreated[a], m.timeCreated[b] = m.timeCreated[b], m.timeCreated[a]
	for d := 0; d < 2; d++ {
		m.node[m.node[a].link[d].b].link[1-d].b = b
	}
	m.node[a], m.node[b] = m.node[b], m.node[a]
}

func (m *B5) setSurfaceAndGroup(b, sf, g int) {
	m.surface[b] = sf
	m.group[b] = g
	if b <= m.n {
		return
	}
	for bb := m.base[b]; m.surface[bb] != sf; bb = m.node[bb].nextB() {
		m.setSurfaceAndGroup(bb, sf, g)
	}
}

func (m *B5) mergeSmallerBlossoms(bid int) {
	lb := bid
	largestSize := 1
	beta := m.base[bid]
	b := beta
	for {
		if m.node[b].size > largestSize {
			largestSize = m.node[b].size
			lb = b
		}
		b = m.node[b].nextB()
		if b == beta {
			break
		}
	}
	beta = m.base[bid]
	b = beta
	for {
		if b != lb {
			m.setSurfaceAndGroup(b, lb, b)
		}
		b = m.node[b].nextB()
		if b == beta {
			break
		}
	}
	m.group[lb] = lb
	if largestSize > 1 {
		m.surface[bid], m.heavy[bid] = lb, lb
		m.swapBlossom(lb, bid)
	} else {
		m.heavy[bid] = 0
	}
}

func (m *B5) contract(x, y, eid int) {
	bx := m.surface[x]
	by := m.surface[y]
	if bx == by {
		panic("bx == by")
	}
	h := -(eid + 1)
	m.link[m.surface[m.mate[bx]]].from = h
	m.link[m.surface[m.mate[by]]].from = h

	lca := -1
	for {
		if m.mate[by] != 0 {
			bx, by = by, bx
		}
		lca = m.surface[m.link[bx].from]
		bx = lca
		if m.link[m.surface[m.mate[bx]]].from == h {
			break
		}
		m.link[m.surface[m.mate[bx]]].from = h
	}

	m.unusedBidIdx--
	bid := m.unusedBid[m.unusedBidIdx]
	if m.unusedBidIdx < 0 {
		panic("unusedBidIdx < 0")
	}
	treeSize := 0
	for d := 0; d < 2; d++ {
		for bv := m.surface[x]; bv != lca; {
			mv := m.mate[bv]
			bmv := m.surface[mv]
			v := m.mate[mv]
			f, t := m.link[v].from, m.link[v].to
			treeSize += m.node[bv].size + m.node[bmv].size
			m.link[mv] = link{x, y}

			if bv > m.n {
				m.potential[bv] += (m.timeCurrent - m.timeCreated[bv]) << 1
			}
			if bmv > m.n {
				m.heap4.Erase(bmv)
			}
			m.pushOuterAndFixPotentials(bmv, m.fixBlossomPotential(bmv, kInner))

			m.node[bv].link[d] = nodeLink{bmv, mv}
			m.node[bmv].link[1-d] = nodeLink{bv, v}
			bv = m.surface[f]
			m.node[bmv].link[d] = nodeLink{bv, f}
			m.node[bv].link[1-d] = nodeLink{bmv, t}
		}
		m.node[m.surface[x]].link[1-d] = nodeLink{m.surface[y], y}
		x, y = y, x
	}
	if lca > m.n {
		m.potential[lca] += (m.timeCurrent - m.timeCreated[lca]) << 1
	}
	m.node[bid].size = treeSize + m.node[lca].size
	m.base[bid] = lca
	m.link[bid] = m.link[lca]
	m.mate[bid] = m.mate[lca]
	m.label[bid] = kOuter
	m.surface[bid] = bid
	m.timeCreated[bid] = m.timeCurrent
	m.potential[bid] = 0
	m.lazy[bid] = 0
	m.mergeSmallerBlossoms(bid)
}

func (m *B5) linkBlossom(v int, l link) {
	m.link[v] = l
	if v <= m.n {
		return
	}
	b := m.base[v]
	m.linkBlossom(b, l)
	pb := m.node[b].prevB()
	l = link{m.node[pb].nextV(), m.node[b].prevV()}
	bv := b
	for {
		bw := m.node[bv].nextB()
		if bw == b {
			break
		}
		m.linkBlossom(bw, l)
		nl := link{m.node[bw].prevV(), m.node[bv].nextV()}
		bv = m.node[bw].nextB()
		m.linkBlossom(bv, nl)
	}
}

func (m *B5) pushOuterAndFixPotentials(v, d int) {
	m.label[v] = kOuter
	if v > m.n {
		for b := m.base[v]; m.label[b] != kOuter; b = m.node[b].nextB() {
			m.pushOuterAndFixPotentials(b, d)
		}
	} else {
		m.potential[v] += m.timeCurrent + d
		if m.potential[v] < m.event1.time {
			m.event1 = event{time: m.potential[v], id: v}
		}
		m.push(v)
	}
}

func (m *B5) grow(root, x, y int) bool {
	by := m.surface[y]
	visited := m.label[by] != kFree
	if !visited {
		m.linkBlossom(by, link{0, 0})
	}
	m.label[by] = kInner
	m.timeCreated[by] = m.timeCurrent
	m.heap2.Erase(by)
	if y != by {
		m.heap4.Update(by, m.timeCurrent+(m.potential[by]>>1))
	}
	z := m.mate[by]
	if z == 0 && by != m.surface[root] {
		m.rematch(x, y)
		m.rematch(y, x)
		return true
	}
	bz := m.surface[z]
	if !visited {
		m.linkBlossom(bz, link{x, y})
	} else {
		m.link[bz] = link{x, y}
		m.link[z] = link{x, y}
	}
	m.pushOuterAndFixPotentials(bz, m.fixBlossomPotential(bz, kFree))
	m.timeCreated[bz] = m.timeCurrent
	m.heap2.Erase(bz)
	return false
}

func (m *B5) freeBlossom(bid int) {
	m.unusedBid[m.unusedBidIdx] = bid
	m.unusedBidIdx++
	m.base[bid] = bid
}

func (m *B5) recalculateMinimumSlack(b, g int) int {
	if b <= m.n {
		if m.slack[b] >= m.slack[g] {
			return 0
		}
		m.slack[g] = m.slack[b]
		m.bestFrom[g] = m.bestFrom[b]
		return b
	}
	v := 0
	beta := m.base[b]
	bb := beta
	for {
		w := m.recalculateMinimumSlack(bb, g)
		if w != 0 {
			v = w
		}
		bb = m.node[bb].nextB()
		if bb == beta {
			break
		}
	}
	return v
}

func (m *B5) constructSmallerComponents(b, sf, g int) {
	m.surface[b] = sf
	m.group[b] = g
	if b <= m.n {
		return
	}
	for bb := m.base[b]; m.surface[bb] != sf; bb = m.node[bb].nextB() {
		if bb == m.heavy[b] {
			m.constructSmallerComponents(bb, sf, g)
		} else {
			m.setSurfaceAndGroup(bb, sf, bb)
			to := 0
			if bb > m.n {
				m.slack[bb] = inf
				to = m.recalculateMinimumSlack(bb, bb)
			} else if m.slack[bb] < inf {
				to = bb
			}
			if to > 0 {
				m.heap2s.Push(sf, bb, edgeEvent{time: m.slack[bb], from: m.bestFrom[bb], to: to})
			}
		}
	}
}

func (m *B5) moveToLargestBlossom(bid int) {
	h := m.heavy[bid]
	d := (m.timeCurrent - m.timeCreated[bid]) + m.lazy[bid]
	m.lazy[bid] = 0
	beta := m.base[bid]
	for b := beta; ; {
		m.timeCreated[b] = m.timeCurrent
		m.lazy[b] = d
		if b != h {
			m.constructSmallerComponents(b, b, b)
			m.heap2s.Erase(bid, b)
		}
		b = m.node[b].nextB()
		if b == beta {
			break
		}
	}
	if h > 0 {
		m.swapBlossom(h, bid)
		bid = h
	}
	m.freeBlossom(bid)
}

func (m *B5) expand(bid int) {
	mv := m.mate[m.base[bid]]
	m.moveToLargestBlossom(bid)
	oldLink := m.link[mv]
	oldBase := m.surface[m.mate[mv]]
	root := m.surface[oldLink.to]
	d := or(m.mate[root] == m.node[root].link[0].v, 1, 0)
	for b := m.node[oldBase].link[1-d].b; b != root; {
		m.label[b] = kSeparated
		m.activateHeap2Node(b)
		b = m.node[b].link[1-d].b
		m.label[b] = kSeparated
		m.activateHeap2Node(b)
		b = m.node[b].link[1-d].b
	}
	for b := oldBase; ; b = m.node[b].link[d].b {
		m.label[b] = kInner
		nb := m.node[b].link[d].b
		if b == root {
			m.link[m.mate[b]] = oldLink
		} else {
			m.link[m.mate[b]] = link{from: m.node[b].link[d].v, to: m.node[nb].link[1-d].v}
		}
		m.link[m.surface[m.mate[b]]] = m.link[m.mate[b]]
		if b > m.n {
			if m.potential[b] == 0 {
				m.expand(b)
			} else {
				m.heap4.Push(b, m.timeCurrent+(m.potential[b]>>1))
			}
		}
		if b == root {
			break
		}
		b = nb
		m.pushOuterAndFixPotentials(nb, m.fixBlossomPotential(nb, kInner))
	}
}

func (m *B5) augment(root int) bool {
	for len(m.que) > 0 {
		x := m.pop()
		bx := m.surface[x]
		if m.potential[x] == m.timeCurrent {
			if x != root {
				m.rematch(x, 0)
			}
			return true
		}
		for eid := m.ofs[x]; eid < m.ofs[x+1]; eid++ {
			e := &m.edges[eid]
			y := e.to
			by := m.surface[y]
			if bx == by {
				continue
			}
			l := m.label[by]
			if l == kOuter {
				t := m.reducedCost(x, y, e) >> 1
				if t == m.timeCurrent {
					m.contract(x, y, eid)
					bx = m.surface[x]
				} else if t < m.event1.time {
					heap.Push(&m.heap3, edgeEvent{time: t, from: x, to: eid})
				}
			} else {
				t := m.reducedCost(x, y, e)
				if t >= inf {
					continue
				}
				if l != kInner {
					if t+m.lazy[by] == m.timeCurrent {
						if m.grow(root, x, y) {
							return true
						}
					} else {
						m.updateHeap2(x, y, by, t, kFree)
					}
				} else if m.mate[x] != y {
					m.updateHeap2(x, y, by, t, kInner)
				}
			}
		}
	}
	return false
}

func (m *B5) adjustDualVariables(root int) bool {
	// rematch
	time1 := m.event1.time
	//grow
	time2 := inf
	if !m.heap2.Empty() {
		time2 = m.heap2.Min().time
	}
	// contract
	time3 := inf
	for len(m.heap3) > 0 {
		e := m.heap3[0]
		x := e.from
		y := m.edges[e.to].to
		if m.surface[x] != m.surface[y] {
			time3 = e.time
			break
		} else {
			heap.Pop(&m.heap3)
		}
	}
	// expand
	time4 := inf
	if !m.heap4.Empty() {
		time4 = m.heap4.Min()
	}
	// events
	timeNext := min(time1, time2, time3, time4)
	if m.timeCurrent > timeNext || timeNext >= inf {
		panic("event")
	}
	m.timeCurrent = timeNext
	if m.timeCurrent == m.event1.time {
		x := m.event1.id
		if x != root {
			m.rematch(x, 0)
		}
		return true
	}
	for !m.heap2.Empty() && m.heap2.Min().time == m.timeCurrent {
		x := m.heap2.Min().from
		y := m.heap2.Min().to
		if m.grow(root, x, y) {
			return true
		}
	}
	for len(m.heap3) > 0 && m.heap3[0].time == m.timeCurrent {
		x := m.heap3[0].from
		eid := m.heap3[0].to
		y := m.edges[eid].to
		heap.Pop(&m.heap3)
		if m.surface[x] == m.surface[y] {
			continue
		}
		m.contract(x, y, eid)
	}
	for !m.heap4.Empty() && m.heap4.Min() == m.timeCurrent {
		b := m.heap4.ArgMin()
		m.heap4.Pop()
		m.expand(b)
	}
	return false
}

func (m *B5) Solve() (matched [][2]int, unmatched []int, weight int) {
	m.initialize()
	m.setPotential()
	//m.findMaximumMatching()
	for u := 1; u <= m.n; u++ {
		if m.mate[u] == 0 {
			m.doEdmondsSearch(u)
		}
	}
	weight = m.computeOptimalValue()
	for u := 1; u <= m.n; u++ {
		if m.mate[u] > u {
			matched = append(matched, [2]int{u, m.mate[u]})
		} else if m.mate[u] == 0 {
			unmatched = append(unmatched, u)
		}
	}
	return
}

func or(i bool, a, b int) int {
	if i {
		return a
	}
	return b
}
