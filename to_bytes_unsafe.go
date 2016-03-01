// +build unsafe

package bloom

import "unsafe"

func toBytes(data string) []byte {
	str := *(*stringHeader)(unsafe.Pointer(&data))
	sl := sliceHeader{
		data: str.data,
		len:  str.len,
		cap:  str.len,
	}
	return *(*[]byte)(unsafe.Pointer(&sl))
}

type stringHeader struct {
	data unsafe.Pointer
	len  int
}

type sliceHeader struct {
	data unsafe.Pointer
	len  int
	cap  int
}
