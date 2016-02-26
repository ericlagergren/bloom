package bloom

import (
	"fmt"
	"math"
)

const (
	word = 64
	mask = 63
)

// Filter is a Bloom filter.
type Filter struct {
	bits   []uint64
	nbits  uint64
	hashes int
}

func (f *Filter) isSet(pos uint64) bool {
	return f.bits[pos/word]&(1<<(pos&mask)) != 0
}

func (f *Filter) set(pos uint64) {
	f.bits[pos/word] |= 1 << (pos & mask)
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
	m := math.Ceil((n0 * math.Log(p)) / lnsq)

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
		// Multiple of word.
		bits:  make([]uint64, nbits/word),
		nbits: nbits,

		// Number rounds of hashing.
		hashes: int(math.Ceil(math.Ln2 * m / n0)),
	}
}

func (f Filter) dump() {
	fmt.Println("%d bits with %d seeds", f.nbits, f.hashes)
}

// Add adds a key to the filter.
func (f *Filter) Add(key string) {
	// "Less Hashing, Same Performance: Building a Better Bloom Filter
	//  by Adam Kirsch and Michael Mitzenmacher"
	// tells us gi(x) = h1(x) + ih2(x).
	a, b := hash(key)
	m := f.nbits
	for i := 0; i < f.hashes; i++ {
		f.set((a + b*uint64(i)) % m)
	}
}

// Has returns true if the key probably exists in the filter.
func (f *Filter) Has(key string) bool {
	a, b := hash(key)
	m := f.nbits
	for i := 0; i < f.hashes; i++ {
		// Any zero bits means the key has not been set yet.
		if !f.isSet((a + b*uint64(i)) % m) {
			return false
		}
	}
	return true
}
