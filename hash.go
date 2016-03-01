package bloom

import "github.com/EricLagergren/siphash"

const (
	k0 = 17697571051839533707
	k1 = 15128385881502100741
)

func hash(data string) (a, b uint64) {
	return siphash.Hash128(k0, k1, toBytes(data))
}

func hash2(data []byte) (a, b uint64) {
	return siphash.Hash128(k0, k1, data)
}
