package helpers

import "github.com/cespare/xxhash/v2"

type Key interface {
	uint64 | string | []byte | byte | int | uint | int32 | uint32 | int64
}

// TODO: Find a way to reuse memhash for the second uint64 hash.
// It is known that padding bytes is unreliable for generating the second hash,
// and also that, although Go has a memhash128 function,
// it cannot be used to generate [2]uint64.
func KeyToHash[K Key](key K) (uint64, uint64) {
	keyAsAny := any(key)
	switch k := keyAsAny.(type) {
	case uint64:
		return k, 0
	case string:
		return MemHashString(k), xxhash.Sum64String(k)
	case []byte:
		return MemHash(k), xxhash.Sum64(k)
	case byte:
		return uint64(k), 0
	case uint:
		return uint64(k), 0
	case int:
		return uint64(k), 0
	case int32:
		return uint64(k), 0
	case uint32:
		return uint64(k), 0
	case int64:
		return uint64(k), 0
	default:
		panic("Key type not supported")
	}
}
