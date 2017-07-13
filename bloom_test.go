package bloom

import (
	"bufio"
	"compress/gzip"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/AndreasBriese/bbloom"
	"github.com/spencerkimball/cbfilter"
	"github.com/willf/bloom"
)

func TestMain(m *testing.M) {
	file, err := os.Open(filepath.Join("_testdata", "words.txt.gz"))
	if err != nil {
		panic(err)
	}
	defer file.Close()

	rc, err := gzip.NewReader(file)
	if err != nil {
		panic(err)
	}
	defer rc.Close()

	s := bufio.NewScanner(rc)
	for i := 0; s.Scan(); i++ {
		if i%2 == 0 {
			if !s.Scan() {
				break
			}
			data = append(data, s.Text())
		} else {
			data = append(data, randString())
		}
	}
	if err := s.Err(); err != nil {
		panic(err)
	}

	filterBloom = New(len(data), prob)
	filterAndreas = bbloom.New(float64(len(data)), prob)
	filterWillf = bloom.NewWithEstimates(uint(len(data)), prob)
	filterSpencer, err = cbfilter.NewFilter(uint32(len(data)), 8, prob)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(data); i += 2 {
		v := data[i]
		filterBloom.Add(v)
		filterAndreas.Add([]byte(v))
		filterWillf.Add([]byte(v))
		filterSpencer.AddKey(v)
		gmap[v] = struct{}{}
	}

	os.Exit(m.Run())
}

const prob = 0.02

func randString() string {
	var p [12]byte
	_, err := rand.Read(p[:])
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", p[:])
}

var (
	data []string
	ghas bool // global 'has'

	filterBloom   *Filter
	filterAndreas bbloom.Bloom
	filterWillf   *bloom.BloomFilter
	filterSpencer *cbfilter.Filter
	gmap          = make(map[string]struct{})
)

func BenchmarkBloom(b *testing.B) {
	var lhas bool
	for i := 0; i < b.N; i++ {
		lhas = filterBloom.Has(data[i%len(data)])
	}
	ghas = lhas
}

func BenchmarkAndreas(b *testing.B) {
	var lhas bool
	for i := 0; i < b.N; i++ {
		lhas = filterAndreas.Has([]byte(data[i%len(data)]))
	}
	ghas = lhas
}

func BenchmarkWillf(b *testing.B) {
	var lhas bool
	for i := 0; i < b.N; i++ {
		lhas = filterWillf.Test([]byte(data[i%len(data)]))
	}
	ghas = lhas
}

func BenchmarkSpencer(b *testing.B) {
	var lhas bool
	for i := 0; i < b.N; i++ {
		lhas = filterSpencer.HasKey(data[i%len(data)])
	}
	ghas = lhas
}

func BenchmarkMap(b *testing.B) {
	var lhas bool
	for i := 0; i < b.N; i++ {
		_, lhas = gmap[data[i%len(data)]]
	}
	ghas = lhas
}

const myFilter = "EricLagergren/bloom"

func TestBloom(t *testing.T) {
	testBloomFunc(t, filterBloom.Has, myFilter)
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

	desc := "GOOD"
	rate := float64(fp) / float64(niters)
	if rate > prob {
		desc = "BAD"
		if name == myFilter {
			t.Fatalf("wanted %f error rate, got %f", prob, rate)
		}
	}
	t.Logf("%s - %d false positives for a rate of: %f (want < %f)",
		desc, fp, rate, prob)
}

func TestFilter_MarshalBinary(t *testing.T) {
	f := New(5, 0.2)
	for _, v := range [...]string{"one", "two", "three", "four", "five"} {
		f.Add(v)
	}
	b, err := f.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	var f2 Filter
	if err = f2.UnmarshalBinary(b); err != nil {
		t.Fatal(err)
	}
}

func TestFilter_Size(t *testing.T) {
	min := func(a, b int) int {
		if a < b {
			return a
		}
		return b
	}
	for size := 1; size < 1e8; size += int(float64(size) * 1.14) {
		f := New(size, 0.2)

		m := min(len(data), size)
		for i := 0; i < m; i++ {
			f.Add(data[i])
		}
		for i := 0; i < (size - m); i++ {
			f.Add(randString())
		}

		sz := f.Size()
		v := float64(size) * 0.05
		max := int(float64(size) + v)
		min := int(float64(size) - v)
		if sz != size && (sz > max || sz < min) {
			t.Fatalf("size(%d): got %d, wanted [%d, %d]", size, sz, min, max)
		}

		if testing.Verbose() {
			t.Logf("Size(%d) ~ %d (=%f)", size, sz, float64(size)/float64(sz))
		}
	}
}
