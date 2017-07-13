package bloom

import (
	"encoding"
	"encoding/binary"
	"errors"
	"math"

	"github.com/ericlagergren/siphash"
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
}

func (f *Filter) isSet(i uint64) bool {
	return ((f.bits[i>>shift] >> (i & (word - 1))) & 1) != 0
}

func (f *Filter) set(i uint64) {
	f.bits[i>>shift] |= 1 << (i & (word - 1))
}

// New creates a new Bloom Filter for n items with probability p.
// p should be the between 0 and 1 and indicate the probability of false
// positives wanted. n should be a positive integer describing the number of
// items in the filter.
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
		bits:   make([]uint64, nbits>>6),
		nbits:  nbits,
		hashes: int(math.Ceil(math.Ln2 * m / n0)), // number of hashing rounds
	}
}

const (
	k0 = 17697571051839533707
	k1 = 15128385881502100741
)

// Add adds a key to the filter.
func (f *Filter) Add(key string) {
	f.AddBytes(toBytes(key))
}

// AddBytes adds a key to the filter.
func (f *Filter) AddBytes(key []byte) {
	// "Less Hashing, Same Performance: Building a Better Bloom Filter
	//  by Adam Kirsch and Michael Mitzenmacher"
	// tells us gi(x) = h1(x) + ih2(x).
	a, b := siphash.Hash128(k0, k1, key)
	m := f.nbits - 1
	for i := 0; i < f.hashes; i++ {
		f.set((a + b*uint64(i)) & m)
	}
}

// Has returns true if the key probably exists in the filter.
func (f *Filter) Has(key string) bool {
	return f.HasBytes(toBytes(key))
}

// HasBytes returns true if the key probably exists in the filter.
func (f *Filter) HasBytes(key []byte) bool {
	a, b := siphash.Hash128(k0, k1, key)
	m := f.nbits - 1
	for i := 0; i < f.hashes; i++ {
		if !f.isSet((a + b*uint64(i)) & m) {
			return false
		}
	}
	return true
}

// Size returns the approximate number of items in the filter. At most it
// should be within 5% of the actual amount, assuming the number of items in
// the filter is <= the original size of the filter.
//
// Full disclosure: in practice, the variance is less than 1%; 5% is absolute
// maximum, tested up to 1e8 elements (see bloom_test.go). The algorithm is
// from http://pubs.acs.org/doi/abs/10.1021/ci600526a, but since I do not have
// access to ACS I do not know if the authors of the authors published the
// variance of the algorithm.
func (f *Filter) Size() int {
	m := float64(f.nbits)
	k := float64(f.hashes)
	X := float64(f.popcount())
	// n* = -(m/k) ln[1 - x/m]
	return -int((m / k) * math.Log(1-(X/m)))
}

func (f *Filter) popcount() int {
	var n int
	for _, word := range f.bits {
		n += popcount(word)
	}
	return n
}

const (
	ver = 1         // marshal version
	bpw = word >> 3 // bytes per word
)

// MarshalBinary implements encoding.BinaryMarshaler.
func (f *Filter) MarshalBinary() (data []byte, err error) {
	data = make([]byte, 0 /* formatting :-) */ +
		/* version */ 1+
		/* hashes  */ bpw+
		/* bits    */ (f.nbits>>3),
	)
	data[0] = ver
	binary.LittleEndian.PutUint64(data[1:], uint64(f.hashes))
	for i, w := range f.bits {
		binary.LittleEndian.PutUint64(data[1+(bpw*(i+1)):], w)
	}
	return data, nil
}

// MarshalBinary implements encoding.BinaryUnmarshaler.
func (f *Filter) UnmarshalBinary(data []byte) error {
	if len(data) < 1+bpw {
		return errors.New("bloom.UnmarshalBinary: data too short, unknown encoding")
	}
	if data[0] != ver {
		return errors.New("bloom.UnmarshalBinary: unknown encoding")
	}
	f.hashes = int(binary.LittleEndian.Uint64(data[1:]))
	data = data[1+bpw:]
	f.nbits = uint64(len(data) << 3)
	f.bits = make([]uint64, (f.nbits >> 6))
	for i := range f.bits {
		f.bits[i] = binary.LittleEndian.Uint64(data[bpw*i:])
	}
	return nil
}

var (
	_ interface {
		encoding.BinaryMarshaler
		encoding.BinaryUnmarshaler
	} = (*Filter)(nil)
)
