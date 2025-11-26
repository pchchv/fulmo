package fulmo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTinyLFUIncrement(t *testing.T) {
	a := newTinyLFU(4)
	a.Increment(1)
	a.Increment(1)
	a.Increment(1)
	require.True(t, a.door.Has(1))
	require.Equal(t, int64(2), a.freq.Estimate(1))

	a.Increment(1)
	require.False(t, a.door.Has(1))
	require.Equal(t, int64(1), a.freq.Estimate(1))
}

func TestTinyLFUEstimate(t *testing.T) {
	a := newTinyLFU(8)
	a.Increment(1)
	a.Increment(1)
	a.Increment(1)
	require.Equal(t, int64(3), a.Estimate(1))
	require.Equal(t, int64(0), a.Estimate(2))
}

func TestTinyLFUPush(t *testing.T) {
	a := newTinyLFU(16)
	a.Push([]uint64{1, 2, 2, 3, 3, 3})
	require.Equal(t, int64(1), a.Estimate(1))
	require.Equal(t, int64(2), a.Estimate(2))
	require.Equal(t, int64(3), a.Estimate(3))
	require.Equal(t, int64(6), a.incrs)
}

func TestTinyLFUClear(t *testing.T) {
	a := newTinyLFU(16)
	a.Push([]uint64{1, 3, 3, 3})
	a.clear()
	require.Equal(t, int64(0), a.incrs)
	require.Equal(t, int64(0), a.Estimate(3))
}

func TestSampledLFUAdd(t *testing.T) {
	e := newSampledLFU(4)
	e.add(1, 1)
	e.add(2, 2)
	e.add(3, 1)
	require.Equal(t, int64(4), e.used)
	require.Equal(t, int64(2), e.keyCosts[2])
}

func TestSampledLFUDel(t *testing.T) {
	e := newSampledLFU(4)
	e.add(1, 1)
	e.add(2, 2)
	e.del(2)
	require.Equal(t, int64(1), e.used)
	_, ok := e.keyCosts[2]
	require.False(t, ok)
	e.del(4)
}

func TestSampledLFUUpdate(t *testing.T) {
	e := newSampledLFU(4)
	e.add(1, 1)
	require.True(t, e.updateIfHas(1, 2))
	require.Equal(t, int64(2), e.used)
	require.False(t, e.updateIfHas(2, 2))
}

func TestSampledLFUClear(t *testing.T) {
	e := newSampledLFU(4)
	e.add(1, 1)
	e.add(2, 2)
	e.add(3, 1)
	e.clear()
	require.Equal(t, 0, len(e.keyCosts))
	require.Equal(t, int64(0), e.used)
}

func TestSampledLFURoom(t *testing.T) {
	e := newSampledLFU(16)
	e.add(1, 1)
	e.add(2, 2)
	e.add(3, 3)
	require.Equal(t, int64(6), e.roomLeft(4))
}

func TestSampledLFUSample(t *testing.T) {
	e := newSampledLFU(16)
	e.add(4, 4)
	e.add(5, 5)
	sample := e.fillSample([]*policyPair{
		{1, 1},
		{2, 2},
		{3, 3},
	})
	k := sample[len(sample)-1].key
	require.Equal(t, 5, len(sample))
	require.NotEqual(t, 1, k)
	require.NotEqual(t, 2, k)
	require.NotEqual(t, 3, k)
	require.Equal(t, len(sample), len(e.fillSample(sample)))
	e.del(5)
	sample = e.fillSample(sample[:len(sample)-2])
	require.Equal(t, 4, len(sample))
}

func TestPolicy(t *testing.T) {
	defer func() {
		require.Nil(t, recover())
	}()
	newPolicy[int](100, 10)
}

func TestPolicyClose(t *testing.T) {
	defer func() {
		require.NotNil(t, recover())
	}()

	p := newDefaultPolicy[int](100, 10)
	p.Add(1, 1)
	p.Close()
	p.itemsCh <- []uint64{1}
}

func TestPushAfterClose(t *testing.T) {
	p := newDefaultPolicy[int](100, 10)
	p.Close()
	require.False(t, p.Push([]uint64{1, 2}))
}

func TestAddAfterClose(t *testing.T) {
	p := newDefaultPolicy[int](100, 10)
	p.Close()
	p.Add(1, 1)
}

func TestPolicyClear(t *testing.T) {
	p := newDefaultPolicy[int](100, 10)
	p.Add(1, 1)
	p.Add(2, 2)
	p.Add(3, 3)
	p.Clear()
	require.Equal(t, int64(10), p.Cap())
	require.False(t, p.Has(1))
	require.False(t, p.Has(2))
	require.False(t, p.Has(3))
}

func TestPolicyUpdate(t *testing.T) {
	p := newDefaultPolicy[int](100, 10)
	p.Add(1, 1)
	p.Update(1, 2)
	p.Lock()
	require.Equal(t, int64(2), p.evict.keyCosts[1])
	p.Unlock()
}

func TestPolicyCap(t *testing.T) {
	p := newDefaultPolicy[int](100, 10)
	p.Add(1, 1)
	require.Equal(t, int64(9), p.Cap())
}

func TestPolicyHas(t *testing.T) {
	p := newDefaultPolicy[int](100, 10)
	p.Add(1, 1)
	require.True(t, p.Has(1))
	require.False(t, p.Has(2))
}

func TestPolicyDel(t *testing.T) {
	p := newDefaultPolicy[int](100, 10)
	p.Add(1, 1)
	p.Del(1)
	p.Del(2)
	require.False(t, p.Has(1))
	require.False(t, p.Has(2))
}

func TestPolicyCost(t *testing.T) {
	p := newDefaultPolicy[int](100, 10)
	p.Add(1, 2)
	require.Equal(t, int64(2), p.Cost(1))
	require.Equal(t, int64(-1), p.Cost(2))
}

func TestPolicyMetrics(t *testing.T) {
	p := newDefaultPolicy[int](100, 10)
	p.CollectMetrics(newMetrics())
	require.NotNil(t, p.metrics)
	require.NotNil(t, p.evict.metrics)
}
