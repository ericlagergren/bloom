package bloom

import (
	"unsafe"

	"github.com/dchest/siphash"
)

const (
	seed  = 0xDEABEEF
	seed2 = 0xCAFEBABE
)

func toBytes(data string) []byte {
	str := *(*stringHeader)(unsafe.Pointer(&data))
	sl := sliceHeader{
		data: str.data,
		len:  str.len,
		cap:  str.len,
	}
	return *(*[]byte)(unsafe.Pointer(&sl))
}

// siphash
func hash2(data string) uint32 {
	b := toBytes(data)
	return uint32((siphash.Hash(seed, seed2, b) >> 32))
}

// murmur3 hash
func hash(data string) uint32 {
	const (
		c1 = 0xcc9e2d51
		c2 = 0x1b873593
	)

	h1 := uint32(seed)

	x := (*stringHeader)(unsafe.Pointer(&data))
	p := x.data
	length := x.len
	n := length / 4

	var k1 uint32
	for i := 0; i < n; i++ {
		p, k1 = read32(p)

		k1 *= c1
		k1 = (k1 << 15) | (k1 >> (32 - 15))
		k1 *= c2

		h1 ^= k1
		h1 = (h1 << 13) | (h1 >> (32 - 13))
		h1 = h1*5 + 0xe6546b64
	}

	tail := data[n*4:]

	k1 = 0
	switch length & 3 {
	case 3:
		k1 ^= uint32(tail[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint32(tail[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint32(tail[0])
		k1 *= c1
		k1 = (k1 << 15) | (k1 >> (32 - 15))
		k1 *= c2
		h1 ^= k1
	}

	return fmix32(h1 ^ uint32(length))
}

func fmix32(h uint32) uint32 {
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16
	return h
}

type stringHeader struct {
	data unsafe.Pointer
	len  int
}

type sliceHeader struct {
	data unsafe.Pointer
	len  int
	cap  int
}

// read32 returns a uint32 found at p and returns p + 4.
func read32(p unsafe.Pointer) (unsafe.Pointer, uint32) {
	return unsafe.Pointer(uintptr(p) + 4), *(*uint32)(p)
}
