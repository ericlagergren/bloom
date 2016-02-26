package bloom

import (
	"fmt"
	"math"
)

// Filter is a Bloom filter.
type Filter struct {
	bits    bitvec
	hashers int
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

	return &Filter{
		// Multiple of 64.
		bits: newBitVec((int(m) + _W/2) & (^(_W - 1))),

		// Number rounds of hashing.
		hashers: int(math.Ln2 * m / n0),
	}
}

func (f Filter) dump() {
	fmt.Println("%d bits with %d seeds", f.bits.len(), f.hashers)
}

// Add adds a key to the filter.
func (f *Filter) Add(key string) {
	// Less Hashing, Same Performance: Building a Better Bloom Filter
	// by Adam Kirsch and Michael Mitzenmacher
	// tells us gi(x) = h1(x) + ih2(x).
	a := hash(key)
	b := hash2(key)
	m := uint32(f.bits.len())
	for i := 0; i < f.hashers; i++ {
		f.bits.set((a + b*uint32(i)) % m)
	}
}

// Has returns true if the key probably exists in the filter.
func (f *Filter) Has(key string) bool {
	a := hash(key)
	b := hash2(key)
	m := uint32(f.bits.len())
	for i := 0; i < f.hashers; i++ {
		// Any zero bits means the key has not been set yet.
		if !f.bits.isSet((a + b*uint32(i)) % m) {
			return false
		}
	}
	return true
}
