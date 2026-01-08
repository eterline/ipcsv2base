package ipbase

import "net/netip"

// interner temporarily holds metadata maps during CSV processing.
type interner struct {
	meta map[netip.Prefix]ipMetadata
	c2c  map[code2bytes]code2bytes
}

// newInterner creates a new interner with initialized maps.
func newInterner() *interner {
	return &interner{
		meta: make(map[netip.Prefix]ipMetadata),
		c2c:  make(map[code2bytes]code2bytes),
	}
}

// Clear releases all internal maps.
func (i *interner) Clear() {
	i.meta = nil
	i.c2c = nil
}

// =============================================

// code2bytes stores a 2-character code as a fixed-size array.
type code2bytes [2]byte

var emptyCode = code2bytes{0, 0}

// makecode2bytes creates a code2bytes from a string. Returns emptyCode if string is too short.
func makecode2bytes(s string) code2bytes {
	if len(s) >= 2 {
		return [2]byte{s[0], s[1]}
	}
	return emptyCode
}

// String converts the 2-byte code to a string.
func (c2b code2bytes) String() string {
	if c2b == emptyCode {
		return ""
	}
	return string(c2b[:])
}

// =============================================

// IPVersion represents the IP address version type.
type IPVersion uint8

const (
	IPv4v6 IPVersion = iota // Both IPv4 and IPv6
	IPv4                    // IPv4 only
	IPv6                    // IPv6 only
)

// IPVersionStr converts a string representation ("v4" or "v6") to IPVersion.
func IPVersionStr(s string) IPVersion {
	switch s {
	case "v4":
		return IPv4
	case "v6":
		return IPv6
	default:
		return IPv4v6
	}
}

// validate checks whether the given IP address matches the version.
func (v IPVersion) validate(ip netip.Addr) bool {
	switch v {
	case IPv4v6:
		return true
	case IPv4:
		return ip.Is4() || ip.Is4In6()
	case IPv6:
		return ip.Is6()
	default:
		return false
	}
}
