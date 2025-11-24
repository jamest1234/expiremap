package timequeue_test

import (
	"container/heap"
	"testing"
	"time"

	"github.com/jamest1234/expiremap/timequeue"
)

func TestTimeQueue(t *testing.T) {
	items := map[string]int{
		"banana": 3, "apple": 2, "pear": 4,
	}

	pq := timequeue.TimeQueue[string]{}

	for value, priority := range items {
		heap.Push(&pq, &timequeue.TimeQueueEntry[string]{
			Value:  value,
			Expiry: time.Now().Add(time.Second * time.Duration(priority)),
		})
	}

	heap.Push(&pq, &timequeue.TimeQueueEntry[string]{
		Value:  "orange",
		Expiry: time.Now().Add(time.Second),
	})

	expectedOrder := []string{
		"orange",
		"apple",
		"banana",
		"pear",
	}

	i := 0

	for pq.Len() > 0 {
		item := heap.Pop(&pq).(*timequeue.TimeQueueEntry[string])
		t.Logf("%v:%s\n", item.Expiry.Unix(), item.Value)

		if item.Value != expectedOrder[i] {
			t.Fatalf("expected %s but got %s", expectedOrder[i], item.Value)
		}
		i++
	}
}
