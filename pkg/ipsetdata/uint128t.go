package ipsetdata

import (
	"encoding/binary"
	"net/netip"

	"go4.org/netipx"
)

// uint128t - Unsigned 128-bit integer represented as two uint64 values (hi, lo).
type uint128t struct {
	hi uint64
	lo uint64
}

/*
ToAddr - Converts uint128 value to netip.Addr.

	IPv4-mapped IPv6 addresses are normalized to IPv4.
*/
func (u uint128t) ToAddr() (a netip.Addr) {
	var b [16]byte
	binary.BigEndian.PutUint64(b[:8], u.hi)
	binary.BigEndian.PutUint64(b[8:], u.lo)

	a = netip.AddrFrom16(b)
	if a.Is4In6() {
		a = netip.AddrFrom4(a.As4())
	}
	return a
}

// Addr2Uint128t - Converts netip.Addr to uint128 representation.
func Addr2Uint128t(a netip.Addr) (u uint128t) {
	as16 := a.As16()
	u.hi = binary.BigEndian.Uint64(as16[:8])
	u.lo = binary.BigEndian.Uint64(as16[8:])
	return u
}

/*
Compare - Compares two uint128 values.

	Returns -1, 0 or 1.
*/
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

// Less - Reports whether u < v.
func (u uint128t) Less(v uint128t) bool {
	return u.hi < v.hi || (u.hi == v.hi && u.lo < v.lo)
}

// IPRangeFromUint128ts - Creates IPRange from uint128 boundaries.
func IPRangeFromUint128ts(a, b uint128t) netipx.IPRange {
	return netipx.IPRangeFrom(a.ToAddr(), b.ToAddr())
}

// rangeUint128t - Inclusive IP range represented by uint128 boundaries.
type rangeUint128t struct {
	start uint128t
	end   uint128t
}

// Contains - Reports whether the value is within [start, end].
func (r rangeUint128t) Contains(u uint128t) bool {
	return !u.Less(r.start) && !r.end.Less(u)
}

// ToIPRange - Converts internal range to netipx.IPRange.
func (r rangeUint128t) ToIPRange() netipx.IPRange {
	return IPRangeFromUint128ts(r.start, r.end)
}
