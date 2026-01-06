package ipcsv2base

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net/netip"
)

// ============

const (
	prefixVectorSize = 24
	fieldVectorSize  = 256
)

// ============

type prefixVector [prefixVectorSize]byte

func Prefix2Vector(p netip.Prefix) (prefixVector, error) {
	var out prefixVector

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

var errInvalidPrefixBits = errors.New("invalid prefix bits")

func (vec prefixVector) Vector2Prefix() (netip.Prefix, error) {
	bits := binary.LittleEndian.Uint32(vec[16:20])
	if bits > 128 {
		return netip.Prefix{}, errInvalidPrefixBits
	}

	var a [16]byte
	copy(a[:], vec[:16])

	return netip.PrefixFrom(netip.AddrFrom16(a), int(bits)), nil
}

// ============

type fieldVector [fieldVectorSize]byte

func StringFieldVector(s string) fieldVector {
	var v fieldVector
	for i := range v {
		v[i] = 0
	}

	n := len(s)
	if n > fieldVectorSize {
		n = fieldVectorSize
	}

	copy(v[:n], s)
	return v
}

func (v fieldVector) String() string {
	n := 0
	for n < fieldVectorSize && v[n] != 0 {
		n++
	}
	return string(v[:n])
}

// ============

type NetworkVersion int

const (
	IPv4IPv6 NetworkVersion = iota
	IPv4
	IPv6
)

// ============

type NetworkCountry struct {
	pfx     prefixVector
	country fieldVector
}

func NewNetworkCountry(network netip.Prefix, code string) NetworkCountry {
	pfx, _ := Prefix2Vector(network)
	return NetworkCountry{
		pfx:     pfx,
		country: StringFieldVector(code),
	}
}

func (r NetworkCountry) selfSize() int {
	return prefixVectorSize + fieldVectorSize
}

func (r NetworkCountry) prefixVec() prefixVector {
	return r.pfx
}

func (r *NetworkCountry) setPrefix(p prefixVector) {
	r.pfx = p
}

func (r NetworkCountry) writeFields(w io.Writer) {
	w.Write(r.country[:])
}

func (r *NetworkCountry) readFields(rd io.Reader) error {
	_, err := rd.Read(r.country[:])
	return err
}

func (r NetworkCountry) Prefix() netip.Prefix {
	pfx, _ := r.pfx.Vector2Prefix()
	addr := pfx.Addr()

	if addr.Is4In6() {
		ip4 := netip.AddrFrom4(addr.As4())
		return netip.PrefixFrom(ip4, pfx.Bits())
	}

	return pfx
}

func (r NetworkCountry) Country() string {
	return r.country.String()
}

func (r NetworkCountry) counrtyCode() [2]byte {
	return [2]byte{r.country[0], r.country[1]}
}
