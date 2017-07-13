// +build go1.8,!go1.9

package bloom

func popcount(x uint64) int {
	var count int
	for x != 0 {
		x &= x - 1
		count++
	}
	return count
}
