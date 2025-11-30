package sim

import (
	"bytes"
	"testing"
)

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

func TestCollection(t *testing.T) {
	s := NewUniform(100)
	c := Collection(s, 100)
	if len(c) != 100 {
		t.Fatal("collection not full")
	}
}

func TestStringCollection(t *testing.T) {
	s := NewUniform(100)
	c := StringCollection(s, 100)
	if len(c) != 100 {
		t.Fatal("string collection not full")
	}
}

func TestParseARC(t *testing.T) {
	s := NewReader(ParseARC, bytes.NewReader([]byte{
		'1', '2', '7', ' ', '6', '4', ' ', '0', ' ', '0', '\r', '\n',
		'1', '9', '1', ' ', '3', '6', ' ', '0', ' ', '0', '\r', '\n',
	}))
	for i := uint64(0); i < 100; i++ {
		if v, err := s(); err != nil {
			t.Fatal(err)
		} else if v != 127+i {
			t.Fatal("value mismatch")
		}
	}
}
