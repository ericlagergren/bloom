// +build go1.9

package bloom

import "math/bits"

func popcount(x uint64) { return bits.OnesCount64(x) }
