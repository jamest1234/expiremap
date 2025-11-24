package expiremap

import (
	"context"
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
