package fulmo

import "time"

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
