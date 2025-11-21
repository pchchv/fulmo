package helpers

type Key interface {
	uint64 | string | []byte | byte | int | uint | int32 | uint32 | int64
}
