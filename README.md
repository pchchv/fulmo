# fulmo [![Godoc Reference](https://pkg.go.dev/badge/github.com/pchchv/fulmo)](https://pkg.go.dev/github.com/pchchv/fulmo) [![Go Report Card](https://goreportcard.com/badge/github.com/pchchv/fulmo)](https://goreportcard.com/report/github.com/pchchv/fulmo)

Fulmo is a fast, concurrent caching package built with performance and correctness in mind.

## Features

- **High Hit Ratios** - Fulmo's unique combination of admission/eviction policies creates its unique performance.
  - **Eviction: SampledLFU** - on par with exact LRU and better performance on Search and Database traces.
  - **Admission: TinyLFU** - extra performance with little memory overhead (12 bits per counter).
- **Fast Throughput** - superior throughput results from the use of various contention management techniques.
- **Cost-Based Eviction** - any new large item that is considered valuable can evict several smaller items (cost could be anything).
- **Fully Concurrent** - it is possible to use any number of goroutines, with little throughput degradation.
- **Metrics** - optional performance metrics for throughput, hit ratios, and other statistics.
- **Simple API** - just figure out ideal `Config` values and it's off and running.

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