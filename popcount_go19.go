// +build go1.9

package bloom

import "math/bits"

func popcount(x uint64) uint64 { return uint64(bits.OnesCount64(x)) }
