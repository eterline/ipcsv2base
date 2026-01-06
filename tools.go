package ipcsv2base

import (
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"io"
	"net/netip"
)

func upperByte(b byte) byte {
	if b >= 'a' && b <= 'z' {
		return b - ('a' - 'A')
	}
	return b
}

func codeToUint16LE(b0, b1 byte) uint16 {
	return uint16(b1)<<8 | uint16(b0)
}

func resolveCsvFields(r io.Reader, fields ...string) ([]int, error) {
	csvR := csv.NewReader(r)

	header, err := csvR.Read()
	if err != nil {
		return nil, err
	}

	index := make(map[string]int, len(header))
	for i, name := range header {
		index[name] = i
	}

	out := make([]int, 0, len(fields))
	for _, f := range fields {
		i, ok := index[f]
		if !ok {
			return nil, fmt.Errorf("csv field %q not found", f)
		}
		out = append(out, i)
	}

	return out, nil
}

func PrefixToVec24(p netip.Prefix) ([24]byte, error) {
	var out [24]byte

	if !p.IsValid() {
		return out, fmt.Errorf("invalid prefix")
	}

	addr := p.Addr().As16()
	out[0] = addr[0]
	out[1] = addr[1]
	out[2] = addr[2]
	out[3] = addr[3]
	out[4] = addr[4]
	out[5] = addr[5]
	out[6] = addr[6]
	out[7] = addr[7]
	out[8] = addr[8]
	out[9] = addr[9]
	out[10] = addr[10]
	out[11] = addr[11]
	out[12] = addr[12]
	out[13] = addr[13]
	out[14] = addr[14]
	out[15] = addr[15]

	binary.LittleEndian.PutUint32(out[16:20], uint32(p.Bits()))

	return out, nil
}

func Vec24ToPrefix(b [24]byte) (netip.Prefix, error) {
	addr16 := [16]byte{}
	copy(addr16[:], b[0:16])

	bits := int(binary.LittleEndian.Uint32(b[16:20]))

	addr := netip.AddrFrom16(addr16)

	if bits < 0 || bits > 128 {
		return netip.Prefix{}, fmt.Errorf("invalid prefix bits")
	}

	pfx := netip.PrefixFrom(addr, bits)
	if !pfx.IsValid() {
		return netip.Prefix{}, fmt.Errorf("invalid prefix")
	}

	return pfx, nil
}
