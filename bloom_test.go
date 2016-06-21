package bloom

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"

	cr "crypto/rand"

	"github.com/AndreasBriese/bbloom"
	"github.com/spencerkimball/cbfilter"
	"github.com/willf/bloom"
)

func init() {
	var err error
	file, err := os.Open("domains.txt")
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	s := bufio.NewScanner(file)
	for s.Scan() {
		data = append(data, s.Text())
	}

	var b [8]byte
	if _, err := cr.Read(b[:]); err != nil {
		log.Fatalln(err)
	}
	rand.Seed(int64(binary.LittleEndian.Uint64(b[:])))
	str = randString()
	buf = []byte(str)

	filterBloom = New(len(data), prob)
	filterAndreas = bbloom.New(float64(len(data)), prob)
	filterWillf = bloom.NewWithEstimates(uint(len(data)), prob)
	filterSpencer, err = cbfilter.NewFilter(uint32(len(data)), 8, prob)
	if err != nil {
		log.Fatalln(err)
	}
	for i := range data {
		filterBloom.Add(data[i])
		filterAndreas.Add([]byte(data[i]))
		filterWillf.Add([]byte(data[i]))
		filterSpencer.AddKey(data[i])
		gmap[data[i]] = struct{}{}
	}
}

const prob = 0.02

func randString() string {
	const strlen = 25
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789=-()!@#$%^&*<>./?:;'`\\|"
	var result [strlen]byte
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result[:])
}

var (
	data []string
	str  string
	buf  []byte
	has  bool

	filterBloom   *Filter
	filterAndreas bbloom.Bloom
	filterWillf   *bloom.BloomFilter
	filterSpencer *cbfilter.Filter
	gmap          = make(map[string]struct{})
)

func BenchmarkBloom(b *testing.B) {
	for i := 0; i < b.N; i++ {
		has = filterBloom.Has(str)
	}
}

func BenchmarkAndreas(b *testing.B) {
	for i := 0; i < b.N; i++ {
		has = filterAndreas.Has(buf)
	}
}

func BenchmarkWillf(b *testing.B) {
	for i := 0; i < b.N; i++ {
		has = filterWillf.Test(buf)
	}
}

func BenchmarkSpencer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		has = filterSpencer.HasKey(str)
	}
}

func BenchmarkMap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, has = gmap[str]
	}
}
func TestBloom(t *testing.T) {
	testBloomFunc(t, filterBloom.Has, "EricLagergren/bloom")
}

func TestAndreas(t *testing.T) {
	testBloomFunc(t, func(s string) bool {
		return filterAndreas.Has([]byte(s))
	}, "AndreasBriese/bbloom")
}

func TestWillf(t *testing.T) {
	testBloomFunc(t, filterWillf.TestString, "willf/bloom")
}

func TestSpencerKimball(t *testing.T) {
	testBloomFunc(t, filterSpencer.HasKey, "spencerkimball/cbfilter")
}

func testBloomFunc(t *testing.T, fn func(string) bool, name string) {
	const niters = 1e6
	var fp int
	for i := 0; i < niters; i++ {
		if fn(randString()) {
			fp++
		}
	}
	err := float64(fp) / float64(niters)
	if err > prob {
		if testing.Verbose() {
			t.Errorf("wanted %f error rate, got %f", prob, err)
		} else {
			fmt.Printf("ERROR: (%s): wanted %f error rate, got %f\n", name, prob, err)
		}
	}
	t.Logf("%d false positives for an error rate of: %f", fp, err)
}

func TestFilter_MarshalBinary(t *testing.T) {
	f := New(5, 0.2)
	for _, v := range [...]string{"one", "two", "three", "four", "five"} {
		f.Add(v)
	}
	b, err := f.MarshalBinary()
	if err != nil {
		log.Fatalln(err)
	}
	var f2 Filter
	err = f2.UnmarshalBinary(b)
	if err != nil {
		log.Fatalln(err)
	}
}
