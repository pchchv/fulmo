package fulmo

import (
	"testing"

	"github.com/pchchv/fulmo/helpers"
	"github.com/stretchr/testify/require"
)

func TestStoreUpdate(t *testing.T) {
	s := newStore[int]()
	key, conflict := helpers.KeyToHash(1)
	i := Item[int]{
		Key:      key,
		Conflict: conflict,
		Value:    1,
	}
	s.Set(&i)
	i.Value = 2
	_, ok := s.Update(&i)
	require.True(t, ok)

	val, ok := s.Get(key, conflict)
	require.True(t, ok)
	require.NotNil(t, val)

	val, ok = s.Get(key, conflict)
	require.True(t, ok)
	require.Equal(t, 2, val)

	i.Value = 3
	_, ok = s.Update(&i)
	require.True(t, ok)

	val, ok = s.Get(key, conflict)
	require.True(t, ok)
	require.Equal(t, 3, val)

	key, conflict = helpers.KeyToHash(2)
	i = Item[int]{
		Key:      key,
		Conflict: conflict,
		Value:    2,
	}
	_, ok = s.Update(&i)
	require.False(t, ok)
	val, ok = s.Get(key, conflict)
	require.False(t, ok)
	require.Empty(t, val)
}

func TestShouldUpdate(t *testing.T) {
	// create a should update function where the value only increases
	s := newStore[int]()
	s.SetShouldUpdateFn(func(cur, prev int) bool {
		return cur > prev
	})

	key, conflict := helpers.KeyToHash(1)
	i := Item[int]{
		Key:      key,
		Conflict: conflict,
		Value:    2,
	}
	s.Set(&i)
	i.Value = 1
	_, ok := s.Update(&i)
	require.False(t, ok)

	i.Value = 3
	_, ok = s.Update(&i)
	require.True(t, ok)
}

func BenchmarkStoreGet(b *testing.B) {
	s := newStore[int]()
	key, conflict := helpers.KeyToHash(1)
	i := Item[int]{
		Key:      key,
		Conflict: conflict,
		Value:    1,
	}
	s.Set(&i)
	b.SetBytes(1)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s.Get(key, conflict)
		}
	})
}

func BenchmarkStoreSet(b *testing.B) {
	s := newStore[int]()
	key, conflict := helpers.KeyToHash(1)
	b.SetBytes(1)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := Item[int]{
				Key:      key,
				Conflict: conflict,
				Value:    1,
			}
			s.Set(&i)
		}
	})
}

func BenchmarkStoreUpdate(b *testing.B) {
	s := newStore[int]()
	key, conflict := helpers.KeyToHash(1)
	i := Item[int]{
		Key:      key,
		Conflict: conflict,
		Value:    1,
	}
	s.Set(&i)
	b.SetBytes(1)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s.Update(&Item[int]{
				Key:      key,
				Conflict: conflict,
				Value:    2,
			})
		}
	})
}
