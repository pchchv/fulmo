package fulmo

import (
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

func newTestCache() (*Cache[int, int], error) {
	return NewCache(&Config[int, int]{
		NumCounters: 100,
		MaxCost:     10,
		BufferItems: 64,
		Metrics:     true,
	})
}
