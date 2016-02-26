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
func hash(data string) (a, b uint64) {
	return siphash.Hash128(seed, seed2, toBytes(data))
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
