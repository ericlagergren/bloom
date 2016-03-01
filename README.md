# bloom
Bloom filter in Go

## What
Package bloom is a fast, space-efficient bloom filter.

## Tips
Install or build this package with `-tags=unsafe` to utilize faster
`string` to `[]byte` conversions.

These conversions are *only* done when `-tags=unsafe` is added and *only*
done inside the hash function.

On my old i3 laptop this increases the speed of the bloom filter from ~75ns to ~64ns.

## Performance
```
BenchmarkAndreas-4	20000000	        75.8 ns/op
BenchmarkBloom-4  	20000000	        68.1 ns/op
BenchmarkWillf-4  	 5000000	       351 ns/op
BenchmarkSpencer-4	 5000000	       314 ns/op
```

## GoDoc
[GoDoc](https://godoc.org/github.com/EricLagergren/bloom)