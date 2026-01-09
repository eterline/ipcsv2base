package ipbase

import (
	"encoding/binary"
	"net/netip"

	"go4.org/netipx"
)

type uint128t struct {
	hi uint64
	lo uint64
}

func (u uint128t) ToAddr() (a netip.Addr) {
	var b [16]byte
	binary.BigEndian.PutUint64(b[:8], u.hi)
	binary.BigEndian.PutUint64(b[8:], u.lo)
	return netip.AddrFrom16(b)
}

func Addr2Uint128t(a netip.Addr) (u uint128t) {
	as16 := a.As16()
	u.hi = binary.BigEndian.Uint64(as16[:8])
	u.lo = binary.BigEndian.Uint64(as16[8:])
	return u
}

func (u uint128t) Compare(v uint128t) int {
	if u.hi < v.hi {
		return -1
	}
	if u.hi > v.hi {
		return 1
	}
	if u.lo < v.lo {
		return -1
	}
	if u.lo > v.lo {
		return 1
	}
	return 0
}

func (u uint128t) Less(v uint128t) bool {
	return u.hi < v.hi || (u.hi == v.hi && u.lo < v.lo)
}

func IPRangeFromUint128ts(a, b uint128t) netipx.IPRange {
	return netipx.IPRangeFrom(a.ToAddr(), b.ToAddr())
}

type rangeUint128t struct {
	start uint128t
	end   uint128t
}

// Contains reports whether u is within [start, end] inclusive.
func (r rangeUint128t) Contains(u uint128t) bool {
	return !u.Less(r.start) && !r.end.Less(u)
}

func (r rangeUint128t) ToIPRange() netipx.IPRange {
	return IPRangeFromUint128ts(r.start, r.end)
}
