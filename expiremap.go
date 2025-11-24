package expiremap

import (
	"container/heap"
	"context"
	"expiremap/timequeue"
	"iter"
	"sync"
	"time"
)

type entry[K comparable, V any] struct {
	key   K
	value V

	timeQueueEntry *timequeue.TimeQueueEntry[K]
}

type Map[K comparable, V any] struct {
	mu sync.Mutex
	m  map[K]*entry[K, V]

	timeQueue     timequeue.TimeQueue[K]
	cleanerSignal chan struct{}
	unloadFunc    func(key K, value V)
}

func New[K comparable, V any](ctx context.Context, unloadFunc func(key K, value V)) *Map[K, V] {
	m := &Map[K, V]{
		m:             map[K]*entry[K, V]{},
		timeQueue:     timequeue.TimeQueue[K]{},
		cleanerSignal: make(chan struct{}, 1),
		unloadFunc:    unloadFunc,
	}
	go m.cleaner(ctx)
	return m
}

func (m *Map[K, V]) cleaner(ctx context.Context) {
	defer m.Clear()

	for {
		m.mu.Lock()
		if m.timeQueue.Len() > 0 {
			nextExpiry := m.timeQueue[0].Expiry
			m.mu.Unlock()
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Until(nextExpiry)):
				m.mu.Lock()
				if m.timeQueue.Len() > 0 {
					nextEntry := m.timeQueue[0]
					if time.Now().After(nextEntry.Expiry) {
						mapEntry, ok := m.m[nextEntry.Value]
						if ok {
							m.remove(mapEntry)
							if m.unloadFunc != nil {
								go m.unloadFunc(mapEntry.key, mapEntry.value)
							}
						}
					}
				}
				m.mu.Unlock()
			case <-m.cleanerSignal:
				continue
			}

		} else {
			m.mu.Unlock()
			select {
			case <-ctx.Done():
				return
			case <-m.cleanerSignal:
				continue
			}
		}
	}
}

// don't call with m.mu locked
func (m *Map[K, V]) wakeCleaner() {
	select {
	case m.cleanerSignal <- struct{}{}:
	default:
	}
}

func (m *Map[K, V]) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.m)
}

func (m *Map[K, V]) Load(key K, newTTL ...time.Duration) (value V, ok bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.m[key]
	if !ok {
		return value, false
	}

	if len(newTTL) > 0 {
		newExpiry := time.Now().Add(newTTL[0])
		if newExpiry.After(entry.timeQueueEntry.Expiry) {
			entry.timeQueueEntry.Expiry = newExpiry
			heap.Fix(&m.timeQueue, entry.timeQueueEntry.Index)
		}
	}

	return entry.value, true
}

func (m *Map[K, V]) Store(key K, value V, ttl time.Duration) {
	m.mu.Lock()

	if e, exists := m.m[key]; exists {
		m.remove(e)
	}

	timeQueueEntry := &timequeue.TimeQueueEntry[K]{
		Value:  key,
		Expiry: time.Now().Add(ttl),
	}

	m.m[key] = &entry[K, V]{
		key:            key,
		value:          value,
		timeQueueEntry: timeQueueEntry,
	}

	heap.Push(&m.timeQueue, timeQueueEntry)

	m.mu.Unlock()
	m.wakeCleaner()
}

func (m *Map[K, V]) remove(e *entry[K, V]) {
	delete(m.m, e.key)
	heap.Remove(&m.timeQueue, e.timeQueueEntry.Index)
}

func (m *Map[K, V]) Remove(key K) bool {
	m.mu.Lock()

	if e, exists := m.m[key]; exists {
		m.remove(e)
		m.mu.Unlock()
		m.wakeCleaner()
		return true
	}

	m.mu.Unlock()
	return false
}

func (m *Map[K, V]) Clear() {
	m.mu.Lock()
	if m.unloadFunc != nil {
		for _, v := range m.m {
			m.unloadFunc(v.key, v.value)
		}
	}
	m.m = map[K]*entry[K, V]{}
	m.timeQueue = timequeue.TimeQueue[K]{}
	m.mu.Unlock()
	m.wakeCleaner()
}

func (m *Map[K, V]) Keys() []K {
	m.mu.Lock()
	defer m.mu.Unlock()
	keys := make([]K, 0, len(m.m))
	for key := range m.m {
		keys = append(keys, key)
	}
	return keys
}

func (m *Map[K, V]) Iter() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		m.mu.Lock()
		defer m.mu.Unlock()
		for k, v := range m.m {
			if !yield(k, v.value) {
				return
			}
		}
	}
}
