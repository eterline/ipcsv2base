package toolkit

import "unsafe"

func BytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

func BytesToUint16BE(b []byte) uint16 {
	return uint16(b[0])<<8 | uint16(b[1])
}

func BytesToUint16LE(b []byte) uint16 {
	return uint16(b[1])<<8 | uint16(b[0])
}

func Uint16ToBytesBE(v uint16, dst []byte) {
	dst[0] = byte(v >> 8)
	dst[1] = byte(v)
}

func Uint16ToBytesLE(v uint16, dst []byte) {
	dst[0] = byte(v)
	dst[1] = byte(v >> 8)
}
