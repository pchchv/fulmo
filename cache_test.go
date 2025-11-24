package fulmo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
