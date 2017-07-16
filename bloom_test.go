package bloom

import (
	"bufio"
	"compress/gzip"
	"crypto/rand"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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
	for i := 0; ; i++ {
		if i%2 == 0 {
			if !s.Scan() {
				break
			}
			data = append(data, s.Text())
		} else {
			data = append(data, randString())
		}
	}
	if err = s.Err(); err != nil {
		panic(err)
	}

	filterBloom = New(len(data), prob)
	dynamicBloom = NewDynamic(prob)
	filterAndreas = bbloom.New(float64(len(data)), prob)
	filterWillf = bloom.NewWithEstimates(uint(len(data)), prob)
	filterSpencer, err = cbfilter.NewFilter(uint32(len(data)), 8, prob)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(data); i += 2 {
		v := data[i]
		filterBloom.Add(v)
		dynamicBloom.Add(v)
		filterAndreas.Add([]byte(v))
		filterWillf.Add([]byte(v))
		filterSpencer.AddKey(v)
		gmap[v] = struct{}{}
	}

	os.Exit(m.Run())
}

const prob = 0.0002

func randString() string {
	var p [10]byte
	_, err := rand.Read(p[:])
	if err != nil {
		panic(err)
	}
	return string(p[:])
}

var (
	data []string
	ghas bool // global 'has'

	filterBloom   *Filter
	dynamicBloom  *Dynamic
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

func BenchmarkDynamicBloom(b *testing.B) {
	var lhas bool
	for i := 0; i < b.N; i++ {
		lhas = dynamicBloom.Has(data[i%len(data)])
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

var testInput = []string{"one", "two", "three", "four", "five"}

func TestBloom_Union(t *testing.T) {
	n := len(testInput)
	f1 := New(n, prob)
	f2 := New(n, prob)

	for _, v := range testInput {
		f2.Add(v)
	}

	if err := f1.Union(f1, f2); err != nil {
		t.Fatal(err)
	}

	for _, v := range testInput {
		if !f1.Has(v) || !f2.Has(v) {
			t.Fatalf("both should have %s", v)
		}
	}

	f2.Add("foo")

	if f1.Size() == f2.Size() {
		t.Fatal("f1 and f2 should have different sizes")
	}

	var f Filter
	if err := f.Union(f1, f2); err != nil {
		t.Fatal(err)
	}

	for _, v := range testInput {
		if !f.Has(v) || !f1.Has(v) || !f2.Has(v) {
			t.Fatalf("all should have %s", v)
		}
	}

	if !f.Has("foo") {
		t.Fatal("f must have 'foo'")
	}
}

func TestBloom_Intersect(t *testing.T) {
	n := len(testInput)
	f1 := New(n, prob)
	f2 := New(n, prob)

	for _, v := range testInput {
		f2.Add(v)
	}

	if err := f1.Intersect(f1, f2); err != nil {
		t.Fatal(err)
	}

	for _, v := range testInput {
		if f1.Has(v) {
			t.Fatalf("f1 should not have %s", v)
		}
		f1.Add(v)
	}

	f1.Add("foo")

	var f Filter
	if err := f.Intersect(f1, f2); err != nil {
		t.Fatal(err)
	}

	for _, v := range testInput {
		if !f.Has(v) || !f1.Has(v) || !f2.Has(v) {
			t.Fatalf("all should have %s", v)
		}
	}

	if f.Has("foo") {
		t.Fatal("f should not have 'foo'")
	}
}

func TestBloom_Jaccard(t *testing.T) {
	const inputs = "01233456789"
	const N = 8
	a := New(N, prob) // 0 1 2 3 4 5 6 7
	b := New(N, prob) // 9 8 7 6 5 4 3 2

	for i := 0; i < N; i++ {
		a.Add(string(inputs[i]))
		b.Add(string(inputs[len(inputs)-(i+1)]))
	}

	idx, err := Jaccard(a, b)
	if err != nil {
		t.Fatal(err)
	}
	const c = 4.0 /* intersection */ / 10 /* union */
	if idx != c {
		t.Fatalf("wanted %f, got %f", c, idx)
	}
}

const myFilter = "EricLagergren/bloom."

func TestBloom(t *testing.T) {
	testBloomFunc(t, filterBloom.Has, myFilter+"Filter")
}

func TestDynamicBloom(t *testing.T) {
	testBloomFunc(t, dynamicBloom.Has, myFilter+"Dynamic")
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

func TestMap(t *testing.T) {
	testBloomFunc(t, func(s string) bool { _, ok := gmap[s]; return ok }, "map")
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
		if strings.HasPrefix(name, myFilter) {
			t.Fatalf("wanted %f error rate, got %f", prob, rate)
		}
	}
	t.Logf("%s - %d false positives for a rate of: %f (want < %f)",
		desc, fp, rate, prob)
}

func TestFilter_MarshalBinary(t *testing.T) {
	f := New(len(testInput), prob)
	f2 := New(len(testInput), prob)
	testMarshalBinary(t, f, f2, testInput)
}

func TestDynamic_MarshalBinary(t *testing.T) {
	f := NewDynamic(prob)
	f2 := NewDynamic(prob)
	testMarshalBinary(t, f, f2, testInput)
}

func testMarshalBinary(
	t *testing.T,
	f interface {
		binMarshaler
		Add(string)
	},
	f2 interface {
		binMarshaler
		Has(string) bool
	},
	input []string,
) {
	for _, v := range input {
		f.Add(v)
	}
	b, err := f.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if err = f2.UnmarshalBinary(b); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(f, f2) {
		t.Fatalf("f != f2\n%#v\n%#v", f, f2)
	}
	for _, v := range input {
		if !f2.Has(v) {
			t.Fatalf("f2.Has(%q) was false", v)
		}
	}
}

func TestFilter_Size(t *testing.T) {
	n := 10
	if testing.Short() {
		n = 1
	}

	for i := 0; i < n; i++ {
		min := func(a, b int) int {
			if a < b {
				return a
			}
			return b
		}
		// assume it's always positive
		round := func(a float64) int { return int(math.Floor(a + 0.5)) }
		for size := 1; size < 1e7; size += int(float64(size) * 1.14) {
			f := New(size, prob)

			m := min(len(data), size)
			for i := 0; i < m; i++ {
				f.Add(data[i])
			}
			for i := 0; i < (size - m); i++ {
				f.Add(randString())
			}

			sz := f.Size()
			v := float64(size) * 0.035
			max := round(float64(size) + v)
			min := round(float64(size) - v)
			if sz != size && (sz > max || sz < min) {
				t.Fatalf("size(%d): got %d (%f), wanted [%d, %d]",
					size, sz, 1-(float64(size)/float64(sz)), min, max)
			}

			if testing.Verbose() {
				t.Logf("Size(%d) ~ %d (=%f)", size, sz, float64(size)/float64(sz))
			}
		}
	}
}
