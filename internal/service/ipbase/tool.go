package ipbase

import (
	"net/netip"

	"github.com/eterline/ipcsv2base/internal/model"
)

func NetworkTypeFromAddrWithSubnet(addr netip.Addr) (t model.NetworkType, p netip.Prefix) {
	// Loopback
	if addr.IsLoopback() {
		return model.NetworkLoopback, loopbackPrefix(addr)
	}

	// Test / documentation
	if pfx, ok := testPrefix(addr); ok {
		return model.NetworkTest, pfx
	}

	// Private
	if addr.IsPrivate() {
		return model.NetworkPrivate, privatePrefix(addr)
	}

	// Global
	if addr.IsGlobalUnicast() {
		return model.NetworkGlobal, netip.Prefix{}
	}

	// Unknown
	return model.NetworkUnknown, netip.Prefix{}
}

func loopbackPrefix(addr netip.Addr) netip.Prefix {
	if addr.Is4() {
		return netip.MustParsePrefix("127.0.0.0/8")
	}
	return netip.MustParsePrefix("::1/128")
}

func testPrefix(addr netip.Addr) (netip.Prefix, bool) {
	if addr.Is4() {
		ip := addr.As4()
		switch {
		case ip[0] == 192 && ip[1] == 0 && ip[2] == 2:
			return netip.MustParsePrefix("192.0.2.0/24"), true
		case ip[0] == 198 && ip[1] == 51 && ip[2] == 100:
			return netip.MustParsePrefix("198.51.100.0/24"), true
		case ip[0] == 203 && ip[1] == 0 && ip[2] == 113:
			return netip.MustParsePrefix("203.0.113.0/24"), true
		}
	} else if addr.Is6() {
		ip := addr.As16()
		if ip[0] == 0x20 && ip[1] == 0x01 && ip[2] == 0x0d && ip[3] == 0xb8 {
			return netip.MustParsePrefix("2001:db8::/32"), true
		}
	}
	return netip.Prefix{}, false
}

func privatePrefix(addr netip.Addr) netip.Prefix {
	if addr.Is4() {
		ip := addr.As4()
		switch {
		case ip[0] == 10:
			return netip.MustParsePrefix("10.0.0.0/8")
		case ip[0] == 172 && ip[1]&0xf0 == 16:
			return netip.MustParsePrefix("172.16.0.0/12")
		case ip[0] == 192 && ip[1] == 168:
			return netip.MustParsePrefix("192.168.0.0/16")
		}
	} else if addr.Is6() {
		// Unique Local Addresses
		if addr.As16()[0]&0xfe == 0xfc {
			return netip.MustParsePrefix("fc00::/7")
		}
	}
	return netip.PrefixFrom(addr, addr.BitLen())
}
