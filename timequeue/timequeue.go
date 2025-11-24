package timequeue

import (
	"time"
)

type TimeQueueEntry[V any] struct {
	Value V

	Expiry time.Time
	Index  int
}

type TimeQueue[V any] []*TimeQueueEntry[V]

func (pq TimeQueue[V]) Len() int { return len(pq) }

func (pq TimeQueue[V]) Less(i, j int) bool {
	return pq[i].Expiry.Before(pq[j].Expiry)
}

func (pq TimeQueue[V]) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *TimeQueue[V]) Push(x any) {
	n := len(*pq)
	item := x.(*TimeQueueEntry[V])
	item.Index = n
	*pq = append(*pq, item)
}

func (pq *TimeQueue[V]) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.Index = -1
	*pq = old[0 : n-1]
	return item
}

/* func (pq *TimeQueue[V]) update(item *TimeQueueEntry[V], value V, ttl time.Duration) {
	item.Value = value
	item.Expiry = time.Now().Add(ttl)
	heap.Fix(pq, item.Index)
} */
