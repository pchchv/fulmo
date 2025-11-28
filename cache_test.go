package fulmo

import (
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pchchv/fulmo/helpers"
	"github.com/stretchr/testify/require"
)

var wait = time.Millisecond * 10

func TestMetrics(t *testing.T) {
	newMetrics()
}

func TestNilMetrics(t *testing.T) {
	var m *Metrics
	for _, f := range []func() uint64{
		m.Hits,
		m.Misses,
		m.KeysAdded,
		m.KeysEvicted,
		m.CostEvicted,
		m.SetsDropped,
		m.SetsRejected,
		m.GetsDropped,
		m.GetsKept,
	} {
		require.Equal(t, uint64(0), f())
	}
}

func TestMetricsAddGet(t *testing.T) {
	m := newMetrics()
	m.add(hit, 1, 1)
	m.add(hit, 2, 2)
	m.add(hit, 3, 3)
	require.Equal(t, uint64(6), m.Hits())

	m = nil
	m.add(hit, 1, 1)
	require.Equal(t, uint64(0), m.Hits())
}

func TestMetricsRatio(t *testing.T) {
	m := newMetrics()
	require.Equal(t, float64(0), m.Ratio())

	m.add(hit, 1, 1)
	m.add(hit, 2, 2)
	m.add(miss, 1, 1)
	m.add(miss, 2, 2)
	require.Equal(t, 0.5, m.Ratio())

	m = nil
	require.Equal(t, float64(0), m.Ratio())
}

func TestMetricsString(t *testing.T) {
	m := newMetrics()
	m.add(hit, 1, 1)
	m.add(miss, 1, 1)
	m.add(keyAdd, 1, 1)
	m.add(keyUpdate, 1, 1)
	m.add(keyEvict, 1, 1)
	m.add(costAdd, 1, 1)
	m.add(costEvict, 1, 1)
	m.add(dropSets, 1, 1)
	m.add(rejectSets, 1, 1)
	m.add(dropGets, 1, 1)
	m.add(keepGets, 1, 1)
	require.Equal(t, uint64(1), m.Hits())
	require.Equal(t, uint64(1), m.Misses())
	require.Equal(t, 0.5, m.Ratio())
	require.Equal(t, uint64(1), m.KeysAdded())
	require.Equal(t, uint64(1), m.KeysUpdated())
	require.Equal(t, uint64(1), m.KeysEvicted())
	require.Equal(t, uint64(1), m.CostAdded())
	require.Equal(t, uint64(1), m.CostEvicted())
	require.Equal(t, uint64(1), m.SetsDropped())
	require.Equal(t, uint64(1), m.SetsRejected())
	require.Equal(t, uint64(1), m.GetsDropped())
	require.Equal(t, uint64(1), m.GetsKept())

	require.NotEqual(t, 0, len(m.String()))

	m = nil
	require.Equal(t, 0, len(m.String()))
	require.Equal(t, "unidentified", stringFor(doNotUse))
}

func TestNewCache(t *testing.T) {
	_, err := NewCache(&Config[int, int]{
		NumCounters: 0,
	})
	require.Error(t, err)

	_, err = NewCache(&Config[int, int]{
		NumCounters: 100,
		MaxCost:     0,
	})
	require.Error(t, err)

	_, err = NewCache(&Config[int, int]{
		NumCounters: 100,
		MaxCost:     10,
		BufferItems: 0,
	})
	require.Error(t, err)

	c, err := NewCache(&Config[int, int]{
		NumCounters: 100,
		MaxCost:     10,
		BufferItems: 64,
		Metrics:     true,
	})
	require.NoError(t, err)
	require.NotNil(t, c)
}

func TestCacheGet(t *testing.T) {
	c, err := NewCache(&Config[int, int]{
		NumCounters:        100,
		MaxCost:            10,
		BufferItems:        64,
		IgnoreInternalCost: true,
		Metrics:            true,
	})
	require.NoError(t, err)

	key, conflict := helpers.KeyToHash(1)
	i := Item[int]{
		Key:      key,
		Conflict: conflict,
		Value:    1,
	}
	c.storedItems.Set(&i)
	val, ok := c.Get(1)
	require.True(t, ok)
	require.NotNil(t, val)

	val, ok = c.Get(2)
	require.False(t, ok)
	require.Zero(t, val)

	// 0.5 and not 1.0 because we tried Getting each item twice
	require.Equal(t, 0.5, c.Metrics.Ratio())

	c = nil
	val, ok = c.Get(0)
	require.False(t, ok)
	require.Zero(t, val)
}

func TestCacheMaxCost(t *testing.T) {
	charset := "abcdefghijklmnopqrstuvwxyz0123456789"
	key := func() []byte {
		k := make([]byte, 2)
		for i := range k {
			k[i] = charset[rand.Intn(len(charset))]
		}
		return k
	}
	c, err := NewCache(&Config[[]byte, string]{
		NumCounters: 12960, // 36^2 * 10
		MaxCost:     1e6,   // 1mb
		BufferItems: 64,
		Metrics:     true,
	})
	require.NoError(t, err)
	stop := make(chan struct{}, 8)
	for i := 0; i < 8; i++ {
		go func() {
			for {
				select {
				case <-stop:
					return
				default:
					time.Sleep(time.Millisecond)

					k := key()
					if _, ok := c.Get(k); !ok {
						val := ""
						if rand.Intn(100) < 10 {
							val = "test"
						} else {
							val = strings.Repeat("a", 1000)
						}
						c.Set(key(), val, int64(2+len(val)))
					}
				}
			}
		}()
	}

	for i := 0; i < 20; i++ {
		time.Sleep(time.Second)
		cacheCost := c.Metrics.CostAdded() - c.Metrics.CostEvicted()
		t.Logf("total cache cost: %d\n", cacheCost)
		require.True(t, float64(cacheCost) <= float64(1e6*1.05))
	}

	for i := 0; i < 8; i++ {
		stop <- struct{}{}
	}
}

func TestCacheMetrics(t *testing.T) {
	c, err := NewCache(&Config[int, int]{
		NumCounters:        100,
		MaxCost:            10,
		IgnoreInternalCost: true,
		BufferItems:        64,
		Metrics:            true,
	})
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		c.Set(i, i, 1)
	}
	time.Sleep(wait)
	m := c.Metrics
	require.Equal(t, uint64(10), m.KeysAdded())
}

func TestCacheWithTTL(t *testing.T) {
	// there may be a race condition,
	// so run the test multiple times
	const try = 10
	for i := 0; i < try; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			c, err := NewCache(&Config[int, int]{
				NumCounters: 100,
				MaxCost:     1000,
				BufferItems: 64,
				Metrics:     true,
			})

			require.NoError(t, err)

			// set initial value for key = 1
			insert := c.SetWithTTL(1, 1, 1, 800*time.Millisecond)
			require.True(t, insert)
			time.Sleep(100 * time.Millisecond)

			// get value from cache for key = 1
			val, ok := c.Get(1)
			require.True(t, ok)
			require.NotNil(t, val)
			require.Equal(t, 1, val)

			time.Sleep(1200 * time.Millisecond)

			val, ok = c.Get(1)
			require.False(t, ok)
			require.Zero(t, val)
		})
	}
}

func TestCacheInternalCost(t *testing.T) {
	c, err := NewCache(&Config[int, int]{
		NumCounters: 100,
		MaxCost:     10,
		BufferItems: 64,
		Metrics:     true,
	})
	require.NoError(t, err)

	// get should return false because the cache's cost is
	// too small to storedItems the item when accounting for the internal cost
	c.SetWithTTL(1, 1, 1, 0)
	time.Sleep(wait)
	_, ok := c.Get(1)
	require.False(t, ok)
}

func TestRecacheWithTTL(t *testing.T) {
	c, err := NewCache(&Config[int, int]{
		NumCounters:        100,
		MaxCost:            10,
		IgnoreInternalCost: true,
		BufferItems:        64,
		Metrics:            true,
	})

	require.NoError(t, err)

	// set initial value for key = 1
	insert := c.SetWithTTL(1, 1, 1, 5*time.Second)
	require.True(t, insert)
	time.Sleep(2 * time.Second)

	// get value from cache for key = 1
	val, ok := c.Get(1)
	require.True(t, ok)
	require.NotNil(t, val)
	require.Equal(t, 1, val)

	// wait for expiration
	time.Sleep(5 * time.Second)

	// cached value for key = 1 should be gone
	val, ok = c.Get(1)
	require.False(t, ok)
	require.Zero(t, val)

	// set new value for key = 1
	insert = c.SetWithTTL(1, 2, 1, 5*time.Second)
	require.True(t, insert)
	time.Sleep(2 * time.Second)

	// get value from cache for key = 1
	val, ok = c.Get(1)
	require.True(t, ok)
	require.NotNil(t, val)
	require.Equal(t, 2, val)
}

func newTestCache() (*Cache[int, int], error) {
	return NewCache(&Config[int, int]{
		NumCounters: 100,
		MaxCost:     10,
		BufferItems: 64,
		Metrics:     true,
	})
}
