package fulmo

const (
	itemNew itemFlag = iota
	itemDelete
	itemUpdate
)

type itemFlag byte
