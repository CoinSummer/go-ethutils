package ethutils

import "unsafe"

const GWEI = 1000000000
const SCALE = 1000000000000000000

func Byte32ToBytes(b32 [32]byte) []byte {
	var bs []byte
	bs = b32[:]
	return bs
}

func BytesToBytes32(s []byte) (a *[32]byte) {
	if len(a) <= len(s) {
		a = (*[len(a)]byte)(unsafe.Pointer(&s[0]))
	}
	return a
}