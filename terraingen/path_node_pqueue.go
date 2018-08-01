package main

type PathNode struct {
	cost   float64
	rank   float64
	cell   *WorldMapCell
	parent *PathNode
	open   bool
	closed bool
	index  int
}

type PathNodePQueue struct {
	q []*PathNode
}

func NewPathNodePQueue(capacity int) *PathNodePQueue {
	return &PathNodePQueue{
		make([]*PathNode, 0, capacity)}
}

func (pq *PathNodePQueue) Clear() {
	pq.q = pq.q[:0]
}

func (pq PathNodePQueue) Len() int {
	return len(pq.q)
}

func (pq PathNodePQueue) Less(i, j int) bool {
	return pq.q[i].rank < pq.q[j].rank
}

func (pq PathNodePQueue) Swap(i, j int) {
	pq.q[i], pq.q[j] = pq.q[j], pq.q[i]
	pq.q[i].index = i
	pq.q[j].index = j
}

func (pq *PathNodePQueue) Push(x interface{}) {
	n := len(pq.q)
	no := x.(*PathNode)
	no.index = n
	pq.q = append(pq.q, no)
}

func (pq *PathNodePQueue) Pop() interface{} {
	old := pq.q
	n := len(old)
	no := old[n-1]
	no.index = -1
	pq.q = old[0 : n-1]
	return no
}
