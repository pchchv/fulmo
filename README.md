# fulmo

## Installing

```sh
go get github.com/pchchv/fulmo
```

## Usage

```go
package main

import (
	"fmt"

	"github.com/pchchv/fulmo"
)

func main() {
	cache, err := fulmo.NewCache(&fulmo.Config[string, string]{
		NumCounters: 1e7,     // number of keys to track frequency of (10M)
		MaxCost:     1 << 30, // maximum cost of cache (1GB)
		BufferItems: 64,      // number of keys per Get buffer
	})
	if err != nil {
		panic(err)
	}
	defer cache.Close()

	// set a value with a cost of 1
	cache.Set("key", "value", 1)
	// wait for value to pass through buffers
	cache.Wait()

	// get value from cache
	if value, found := cache.Get("key"); !found {
		panic("missing value")
	} else {
		fmt.Println(value)
	}

	// del value from cache
	cache.Del("key")
}
```