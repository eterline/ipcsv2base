package model

import (
	"errors"
	"net/netip"
)

type GeoCode string

func NewGeoCode(c string) (GeoCode, error) {
	if len(c) != 2 {
		return "", errors.New("invalid geo code")
	}
	return GeoCode(c), nil
}

func (gc GeoCode) String() string {
	return string(gc)
}

type NetworkType uint8

const (
	NetworkUnknown  NetworkType = iota
	NetworkLoopback             // Loopback
	NetworkGlobal               // WAN
	NetworkPrivate              // private / RFC1918 / ULA
	NetworkTest                 // test / documentation / benchmark
)

func (t NetworkType) String() string {
	switch t {
	case NetworkGlobal:
		return "global"
	case NetworkLoopback:
		return "loopback"
	case NetworkPrivate:
		return "private"
	case NetworkTest:
		return "test"
	default:
		return "unknown"
	}
}

func (t NetworkType) FieldLog() LogField {
	return FieldStringer("network_type", t)
}

type (
	IPMetadata struct {
		Type    NetworkType
		Network netip.Prefix
		Geo     IPGeo
		ASN     IPAS
	}

	IPGeo struct {
		ContinentCode GeoCode
		CountryCode   GeoCode
		CountryName   string
	}

	IPAS struct {
		ASN         int32
		CountryCode GeoCode
		Name        string
		Org         string
		Domain      string
	}
)

func (g IPGeo) FieldsLog() []LogField {
	return []LogField{
		FieldStringer("continent_code", g.ContinentCode),
		FieldStringer("country_code", g.CountryCode),
		FieldString("country_name", g.CountryName),
	}
}

func (as IPAS) FieldsLog() []LogField {
	return []LogField{
		Field("as_number", as.ASN),
		FieldStringer("as_country_code", as.CountryCode),
		FieldString("as_name", as.Name),
		FieldString("as_org", as.Org),
		FieldString("as_domain", as.Domain),
	}
}
