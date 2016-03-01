package bloom

import (
	"encoding/binary"
	"errors"
	"math"
)

const (
	word  = 64
	shift = 6
)

// Filter is a Bloom filter.
type Filter struct {
	bits   []uint64
	nbits  uint64
	hashes int
	items  uint64
}

func (f *Filter) getWord(i uint64) *uint64 {
	return &f.bits[i>>shift]
}

func (f *Filter) isSet(i uint64) bool {
	return ((*f.getWord(i) >> (i & (word - 1))) & 1) != 0
}

func (f *Filter) set(i uint64) {
	*f.getWord(i) |= 1 << (i & (word - 1))
}

// New creates a new Bloom Filter for n items with probability p.
// p should be the between 0 and 1 and indicate the probability
// of false positives wanted.
// n should be a positive integer describing the number of items in
// the filter.
func New(n int, p float64) *Filter {

	// log(1 / pow(2, log(2)))
	const lnsq = -0.480453013918201424667102526326664971730552951594545586866864133623665382259834472199948263443926990932715597661358897481255128413358268503177555294880844290839184664798896404335252423673643658092881230886029639112807153031

	n0 := float64(n)
	m := math.Ceil(n0 * math.Log(p) / lnsq)

	nbits := uint64(m)

	if nbits <= 512 {
		nbits = 512
	} else {
		// Next power of two.
		nbits--
		nbits |= nbits >> 1
		nbits |= nbits >> 2
		nbits |= nbits >> 4
		nbits |= nbits >> 8
		nbits |= nbits >> 16
		nbits |= nbits >> 32
		nbits++
	}

	return &Filter{
		bits:  make([]uint64, nbits>>6),
		nbits: nbits,

		// Number rounds of hashing.
		hashes: int(math.Ceil(math.Ln2 * m / n0)),
	}
}

// Add adds a key to the filter.
func (f *Filter) Add(key string) {
	// "Less Hashing, Same Performance: Building a Better Bloom Filter
	//  by Adam Kirsch and Michael Mitzenmacher"
	// tells us gi(x) = h1(x) + ih2(x).
	a, b := hash(key)
	m := f.nbits - 1
	for i := 0; i < f.hashes; i++ {
		f.set((a + b*uint64(i)) & m)
	}
	f.items++
}

// Add adds a key to the filter.
func (f *Filter) AddBytes(key []byte) {
	a, b := hash2(key)
	m := f.nbits - 1
	for i := 0; i < f.hashes; i++ {
		f.set((a + b*uint64(i)) & m)
	}
	f.items++
}

// Has returns true if the key probably exists in the filter.
func (f *Filter) Has(key string) bool {
	a, b := hash(key)
	m := f.nbits - 1
	for i := 0; i < f.hashes; i++ {
		// Any zero bits means the key has not been set yet.
		if !f.isSet((a + b*uint64(i)) & m) {
			return false
		}
	}
	return true
}

// Has returns true if the key probably exists in the filter.
func (f *Filter) HasBytes(key []byte) bool {
	a, b := hash2(key)
	m := f.nbits - 1
	for i := 0; i < f.hashes; i++ {
		// Any zero bits means the key has not been set yet.
		if !f.isSet((a + b*uint64(i)) & m) {
			return false
		}
	}
	return true
}

const ver = 1
const bpw = word >> 3

// MarshalBinary implements encoding.BinaryMarshaler.
func (f *Filter) MarshalBinary() (data []byte, err error) {
	data = make([]byte, 1+(f.nbits>>3))
	data[0] = ver
	for i, w := range f.bits {
		binary.LittleEndian.PutUint64(data[1+(bpw*i):], w)
	}
	return data, nil
}

// MarshalBinary implements encoding.BinaryUnmarshaler.
func (f *Filter) UnmarshalBinary(data []byte) error {
	if data[0] != ver {
		return errors.New("bloom.UnmarshalBinary: unknown encoding")
	}
	data = data[1:]
	f.nbits = uint64(len(data) << 3)
	f.bits = make([]uint64, (f.nbits >> 6))
	for i := range f.bits {
		f.bits[i] = binary.LittleEndian.Uint64(data[bpw*i:])
	}
	return nil
}
