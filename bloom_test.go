package bloom

import (
	"bufio"
	"encoding/binary"
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

	filterBloom = New(len(data), np)
	filterAndreas = bbloom.New(float64(len(data)), np)
	filterWillf = bloom.NewWithEstimates(uint(len(data)), np)
	filterSpencer, err = cbfilter.NewFilter(uint32(len(data)), 8, np)
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

const np = 0.02
const strlen = 25

func randString() string {
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
	buf := []byte(str)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		has = filterAndreas.Has(buf)
	}
}

func BenchmarkWillf(b *testing.B) {
	buf := []byte(str)
	b.ResetTimer()

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

func TestAndreas(t *testing.T) {
	var fp float64
	for i := 0; i < 1e6; i++ {
		if filterAndreas.Has([]byte(randString())) {
			fp++
		}
	}
	err := fp / float64(len(data))
	if err > np*10 {
		t.Fatalf("wanted %f error rate, got %f", np, err)
	}
	t.Logf("andreas: %f %f", fp, err)
}

func TestBloom(t *testing.T) {
	var fp float64
	for i := 0; i < 1e6; i++ {
		if filterBloom.Has(randString()) {
			fp++
		}
	}
	err := fp / float64(len(data))
	if err > np*10 {
		t.Fatalf("wanted %f error rate, got %f", np, err)
	}
	t.Logf("bloom: %f %f", fp, err)
}

func TestWillf(t *testing.T) {
	var fp float64
	for i := 0; i < 1e6; i++ {
		if filterWillf.Test([]byte(randString())) {
			fp++
		}
	}
	err := fp / float64(len(data))
	if err > np*10 {
		t.Fatalf("wanted %f error rate, got %f", np, err)
	}
	t.Logf("willf: %f %f", fp, err)
}

func TestSpencer(t *testing.T) {
	var fp float64
	for i := 0; i < 1e6; i++ {
		if filterSpencer.HasKey(randString()) {
			fp++
		}
	}
	err := fp / float64(len(data))
	if err > np*10 {
		t.Fatalf("wanted %f error rate, got %f", np, err)
	}
	t.Logf("spencer: %f %f", fp, err)
}
