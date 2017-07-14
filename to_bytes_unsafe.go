// +build unsafe

package bloom

import "unsafe"

func toBytes(data string) []byte {
	type sliceHeader struct {
		data unsafe.Pointer
		len  int
		cap  int
	}
	type stringHeader struct {
		data unsafe.Pointer
		len  int
	}
	str := *(*stringHeader)(unsafe.Pointer(&data))
	sl := sliceHeader{
		data: str.data,
		len:  str.len,
		cap:  str.len,
	}
	return *(*[]byte)(unsafe.Pointer(&sl))
}
