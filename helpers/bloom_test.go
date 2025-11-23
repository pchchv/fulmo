package helpers

import (
	"crypto/rand"
	"fmt"
	"testing"
)

var (
	bf        *Bloom
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

func Test_NumberOfWrongs(t *testing.T) {
	var cnt int
	bf = NewBloomFilter(float64(n*10), float64(7))
	for i := range wordlist1 {
		hash := MemHash(wordlist1[i])
		if !bf.AddIfNotHas(hash) {
			cnt++
		}
	}

	//nolint:lll
	fmt.Printf("Bloomfilter New(7* 2**16, 7) (-> size=%v bit): \n            Check for 'false positives': %v wrong positive 'Has' results on 2**16 entries => %v %%\n", len(bf.bitset)<<6, cnt, float64(cnt)/float64(n))

}

func Benchmark_New(b *testing.B) {
	for r := 0; r < b.N; r++ {
		_ = NewBloomFilter(float64(n*10), float64(7))
	}
}

func Benchmark_Has(b *testing.B) {
	b.ResetTimer()
	for r := 0; r < b.N; r++ {
		for i := range wordlist1 {
			hash := MemHash(wordlist1[i])
			bf.Has(hash)
		}
	}
}

func Benchmark_Clear(b *testing.B) {
	bf = NewBloomFilter(float64(n*10), float64(7))
	for i := range wordlist1 {
		hash := MemHash(wordlist1[i])
		bf.Add(hash)
	}

	b.ResetTimer()
	for r := 0; r < b.N; r++ {
		bf.Clear()
	}
}

func Benchmark_Add(b *testing.B) {
	bf = NewBloomFilter(float64(n*10), float64(7))
	b.ResetTimer()
	for r := 0; r < b.N; r++ {
		for i := range wordlist1 {
			hash := MemHash(wordlist1[i])
			bf.Add(hash)
		}
	}
}
