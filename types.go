package ipcsv2base

import (
	"net/netip"
	"unique"
)

// ============

const (
	pfxSize        = 24
	codeSize       = 2
	companyLineStr = 256

	countryRecSize = pfxSize + codeSize
	companyRecSize = pfxSize + companyLineStr*2
)

// ============

type NetworkVersion int

const (
	IPv4IPv6 NetworkVersion = iota
	IPv4
	IPv6
)

// ============
type Container interface {
	Contains(ip netip.Addr) bool
	ContainsPrefix(p netip.Prefix) bool
}

// ============

type NetworkCountry struct {
	network netip.Prefix
	code    unique.Handle[string]
}

func NewNetworkCountry(network netip.Prefix, code string) NetworkCountry {
	return NetworkCountry{
		network: network,
		code:    unique.Make(code),
	}
}

func (n NetworkCountry) Network() netip.Prefix {
	pfx := n.network
	addr := pfx.Addr()

	if addr.Is4In6() {
		ip4 := netip.AddrFrom4(addr.As4())
		return netip.PrefixFrom(ip4, pfx.Bits())
	}

	return pfx
}

func (n NetworkCountry) Code() string {
	return n.code.Value()
}

// ============

type NetworkCompany struct {
	network netip.Prefix
	name    unique.Handle[string]
	org     unique.Handle[string]
}

func NewNetworkCompany(network netip.Prefix, name, org string) NetworkCompany {
	return NetworkCompany{
		network: network,
		name:    unique.Make(name),
		org:     unique.Make(org),
	}
}

func (n NetworkCompany) Network() netip.Prefix {
	return n.network
}

func (n NetworkCompany) Name() string {
	return n.name.Value()
}

func (n NetworkCompany) Org() string {
	return n.org.Value()
}
