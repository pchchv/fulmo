package sim

import "testing"

func TestZipfian(t *testing.T) {
	s := NewZipfian(1.5, 1, 100)
	m := make(map[uint64]uint64, 100)
	for i := 0; i < 100; i++ {
		k, err := s()
		if err != nil {
			t.Fatal(err)
		}
		m[k]++
	}

	if len(m) == 0 || len(m) == 100 {
		t.Fatal("zipfian not skewed")
	}
}

func TestUniform(t *testing.T) {
	s := NewUniform(100)
	for i := 0; i < 100; i++ {
		if _, err := s(); err != nil {
			t.Fatal(err)
		}
	}
}
