package helpers

import (
	"crypto/rand"
	"fmt"
	"testing"
)

var (
	wordlist1 [][]byte
	n         = 1 << 16
)

func TestMain(m *testing.M) {
	wordlist1 = make([][]byte, n)
	for i := range wordlist1 {
		b := make([]byte, 32)
		_, _ = rand.Read(b)
		wordlist1[i] = b
	}

	fmt.Println("\n###############\nbbloom_test.go")
	fmt.Print("Benchmarks relate to 2**16 OP. --> output/65536 op/ns\n###############\n\n")

	m.Run()
}

func Benchmark_New(b *testing.B) {
	for r := 0; r < b.N; r++ {
		_ = NewBloomFilter(float64(n*10), float64(7))
	}
}
