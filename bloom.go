package bloom

import (
	"encoding"
	"encoding/binary"
	"errors"
	"math"

	"github.com/dchest/siphash"
)

// Dynamic is a Bloom filter that doesn't need a pre-set size. The idea comes
// from http://gsd.di.uminho.pt/members/cbm/ps/dbloom.pdf
type Dynamic struct {
	fs []*Filter
}

// NewDynamic creates a new Bloom filter for an unbounded number of items with
// probability p. See New for more information.
func NewDynamic(p float64) *Dynamic {
	return &Dynamic{fs: []*Filter{New(4096, p)}}
}

func (d *Dynamic) AddBytes(key []byte) {
	f := d.fs[len(d.fs)-1]
	f.AddBytes(key)

	if uint64(f.Size()) >= f.N {
		n := f.N * 2

		// Try and catch a Dynamic filter that grows too large and give a nice
		// message.
		const maxInt = uint64(int(^uint(0) >> 1)) // largest platform int
		if n < f.N || n >= maxInt {
			panic("bloom.Dynamic: too large")
		}
		d.fs = append(d.fs, New(int(n), f.P*0.85))
	}
}

func (d *Dynamic) Add(key string) { d.AddBytes(toBytes(key)) }

func (d *Dynamic) HasBytes(key []byte) bool {
	for _, f := range d.fs {
		if f.HasBytes(key) {
			return true
		}
	}
	return false
}

func (d *Dynamic) Has(key string) bool { return d.HasBytes(toBytes(key)) }

// Filter is a Bloom filter.
type Filter struct {
	N uint64  // original number of items
	P float64 // original p value

	bits     []uint64 // bit array
	nbits    uint64   // number of usable bits
	hashes   uint64   // number of hash functions
	popcount int      // popcount
}

// New creates a new Bloom filter for n items with probability p.
// p should be the between 0 and 1 and indicate the probability of false
// positives wanted. n should be a positive integer describing the number of
// items in the filter.
func New(n int, p float64) *Filter {
	if n < 0 {
		panic("bloom.New: negative count")
	}

	// Optimal number of bits, m, is
	// m = -(n log(p) / pow(log(2), 2))

	// pow(log(2), 2)
	const lnsq = 0.480453013918201424667102526326664971730552951594545586866864133623665382259834472199948263443926990932715597661358897481255128413358268503177555294880844290839184664798896404335252423673643658092881230886029639112807153031

	n0 := float64(n)
	m := -math.Ceil(n0 * math.Log(p) / lnsq)

	// Rounding up to nearest 64 simplifies our filter at the expense of using
	// slightly more space. Patches are welcome :-)
	nbits := ((uint64(m) + mod64) >> div64) * wordBits
	hashes := uint64(-int(math.Ceil(math.Log(p) / math.Ln2)))

	return &Filter{
		N:      uint64(n),
		P:      p,
		bits:   make([]uint64, nbits>>div64),
		nbits:  nbits,
		hashes: hashes,
	}
}

const (
	k0 = 17697571051839533707
	k1 = 15128385881502100741

	wordBits  = 64               // bits per word
	div64     = 6                // division by 64
	mod64     = wordBits - 1     // remainder mod 64
	div8      = 3                // division by 8
	wordBytes = wordBits >> div8 // bytes per word
)

// Ideally the API would be reversed (e.g. AddBytes calling Add), but compiler
// optimizations and safe use of 'unsafe' dictate we should convert the string
// to []bytes instead.

// AddBytes adds a key to the filter.
func (f *Filter) AddBytes(key []byte) {
	// "Less Hashing, Same Performance: Building a Better Bloom Filter"
	// (Kirsch and Mitzenmacher) tells us gi(x) = h1(x) + ih2(x).
	a, b := siphash.Hash128(k0, k1, key)
	for h := uint64(0); h != f.hashes; h++ {
		i := (a + b*h) % f.nbits
		mask := uint64(1) << (i & mod64)
		if f.bits[i>>div64]&mask == 0 {
			f.bits[i>>div64] |= mask
			f.popcount++
		}
	}
}

// Add adds a key to the filter.
func (f *Filter) Add(key string) { f.AddBytes(toBytes(key)) }

// HasBytes returns true if the key probably exists in the filter.
func (f *Filter) HasBytes(key []byte) bool {
	a, b := siphash.Hash128(k0, k1, key)
	for h := uint64(0); h != f.hashes; h++ {
		i := (a + b*h) % f.nbits
		if (f.bits[i>>div64])&(1<<(i&mod64)) == 0 {
			return false
		}
	}
	return true
}

// Has returns true if the key probably exists in the filter.
func (f *Filter) Has(key string) bool { return f.HasBytes(toBytes(key)) }

// Size returns the approximate number of items in the filter. At most it
// should be within 3.5% of the actual amount, assuming the number of items in
// the filter is <= the original size of the filter.
//
// Full disclosure: in practice, the variance is less than 1%; 3.5% is absolute
// maximum, tested up to 1e7 elements (see bloom_test.go). The algorithm is
// from http://pubs.acs.org/doi/abs/10.1021/ci600526a, but since I do not have
// access to ACS I do not know if the authors of the authors published the
// variance of the algorithm.
func (f *Filter) Size() int {
	m := float64(f.nbits)
	k := float64(f.hashes)
	X := float64(f.popcount)
	// n* = -(m/k) ln[1 - x/m]
	return int(math.Floor(-((m / k) * math.Log(1-(X/m))) + 0.5))
}

// Stats returns basic memory information. hashes is the number of hash
// functions and nbits is the number of usable bits. The total number of bits
// allocated by the filter will be nbits rounded up to the nearest multiple of
// 64.
func (f *Filter) Stats() (hashes, nbits uint64) {
	return f.hashes, f.nbits
}

const ver = 1 // marshal version

// MarshalBinary implements encoding.BinaryMarshaler.
func (f *Filter) MarshalBinary() (data []byte, err error) {
	data = make([]byte, 0 /* formatting :-) */ +
		/* version */ 1+
		/* n       */ wordBytes+
		/* hashes  */ wordBytes+
		/* bits    */ (f.nbits>>div8),
	)
	data[0] = ver
	binary.LittleEndian.PutUint64(data[1:], f.N)
	binary.LittleEndian.PutUint64(data[1+wordBytes:], f.hashes)
	for i, w := range f.bits {
		offset := 1 + ((i + 2) * wordBytes)
		binary.LittleEndian.PutUint64(data[offset:], w)
	}
	return data, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler.
func (f *Filter) UnmarshalBinary(data []byte) error {
	if len(data) < 1+(wordBytes*2) {
		return errors.New("bloom.UnmarshalBinary: data too short, unknown encoding")
	}
	if data[0] != ver {
		return errors.New("bloom.UnmarshalBinary: unknown encoding")
	}
	data = data[1:]

	f.N = binary.LittleEndian.Uint64(data)
	data = data[wordBytes:]

	f.hashes = binary.LittleEndian.Uint64(data)
	data = data[wordBytes:]

	f.nbits = uint64(len(data)) << div8
	f.bits = make([]uint64, f.nbits>>div64)
	for i := range f.bits {
		f.bits[i] = binary.LittleEndian.Uint64(data[i*wordBytes:])
	}
	return nil
}

var (
	_ interface {
		encoding.BinaryMarshaler
		encoding.BinaryUnmarshaler
	} = (*Filter)(nil)
)
