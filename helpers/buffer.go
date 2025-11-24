package helpers

import (
	"errors"
	"log"
)

func assert(b bool) {
	if !b {
		log.Fatalf("%+v", errors.New("assertion failure"))
	}
}
