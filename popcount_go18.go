// +build go1.8,!go1.9

package bloom

func popcount(x uint64) uint64 {
	var count uint64
	for x != 0 {
		x &= x - 1
		count++
	}
	return count
}
