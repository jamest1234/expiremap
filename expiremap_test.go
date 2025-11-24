package expiremap

import (
	"context"
	"maps"
	"math/rand/v2"
	"strconv"
	"testing"
	"time"
)

func TestMap(t *testing.T) {
	m := New(context.Background(), func(key string, value int) {
		t.Log("Removed", key, time.Millisecond*time.Duration(value))
	})

	for range 20 {
		i := rand.IntN(5000)
		m.Store(strconv.Itoa(i), i, time.Millisecond*time.Duration(i))
	}

	time.Sleep(time.Second * 6)
	if m.Len() > 0 {
		t.Fail()
	}
}

func TestIter(t *testing.T) {
	m := New[string, int](context.Background(), nil)
	m.Store("a", 1, time.Second)
	m.Store("b", 2, time.Second)
	m.Store("c", 3, time.Second)
	m.Store("d", 4, time.Second)
	m.Store("e", 5, time.Second)
	m.Store("f", 6, time.Second)

	collected := maps.Collect(m.Iter())
	t.Log(collected)

	if len(collected) != 6 {
		t.Fatalf("len(collected) = %v", len(collected))
	}
}
