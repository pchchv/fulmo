package fulmo

import (
	"sync"
	"time"
)

type updateFn[V any] func(cur, prev V) bool

type storeItem[V any] struct {
	key        uint64
	value      V
	conflict   uint64
	expiration time.Time
}

// store is the interface fulfilled by all hash map implementations in this file.
// Some hash map implementations are better suited for certain data distributions than others,
// so this allows us to abstract that out for use in Ristretto.
//
// Every store is safe for concurrent usage.
type store[V any] interface {
	// Get returns the value associated with the key parameter.
	Get(uint64, uint64) (V, bool)
	// Expiration returns the expiration time for this key.
	Expiration(uint64) time.Time
	// Set adds the key-value pair to the Map or updates the value if it's
	// already present. The key-value pair is passed as a pointer to an
	// item object.
	Set(*Item[V])
	// Del deletes the key-value pair from the Map.
	Del(uint64, uint64) (uint64, V)
	// Update attempts to update the key with a new value and returns true if
	// successful.
	Update(*Item[V]) (V, bool)
	// Cleanup removes items that have an expired TTL.
	Cleanup(policy *defaultPolicy[V], onEvict func(item *Item[V]))
	// Clear clears all contents of the store.
	Clear(onEvict func(item *Item[V]))
	SetShouldUpdateFn(f updateFn[V])
}

type lockedMap[V any] struct {
	sync.RWMutex
	em           *expirationMap[V]
	data         map[uint64]storeItem[V]
	shouldUpdate updateFn[V]
}

func newLockedMap[V any](em *expirationMap[V]) *lockedMap[V] {
	return &lockedMap[V]{
		em:   em,
		data: make(map[uint64]storeItem[V]),
		shouldUpdate: func(cur, prev V) bool {
			return true
		},
	}
}

func (m *lockedMap[V]) Set(i *Item[V]) {
	if i == nil {
		// if item is nil make this Set a no-op
		return
	}

	m.Lock()
	defer m.Unlock()
	item, ok := m.data[i.Key]

	if ok {
		// item existed already
		// is needed to check the conflict key and reject the update if they do not match
		// only after that the expiration map is updated
		if i.Conflict != 0 && (i.Conflict != item.conflict) {
			return
		}

		if m.shouldUpdate != nil && !m.shouldUpdate(i.Value, item.value) {
			return
		}

		m.em.update(i.Key, i.Conflict, item.expiration, i.Expiration)
	} else {
		// value is not in the map already
		// there's no need to return anything
		// simply add the expiration map
		m.em.add(i.Key, i.Conflict, i.Expiration)
	}

	m.data[i.Key] = storeItem[V]{
		key:        i.Key,
		conflict:   i.Conflict,
		value:      i.Value,
		expiration: i.Expiration,
	}
}

func (m *lockedMap[V]) Expiration(key uint64) time.Time {
	m.RLock()
	defer m.RUnlock()
	return m.data[key].expiration
}

func (m *lockedMap[V]) Clear(onEvict func(item *Item[V])) {
	m.Lock()
	defer m.Unlock()
	i := &Item[V]{}
	if onEvict != nil {
		for _, si := range m.data {
			i.Key = si.key
			i.Conflict = si.conflict
			i.Value = si.value
			onEvict(i)
		}
	}
	m.data = make(map[uint64]storeItem[V])
}

func (m *lockedMap[V]) setShouldUpdateFn(f updateFn[V]) {
	m.shouldUpdate = f
}
