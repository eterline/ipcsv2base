package ipbase

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"strconv"

	"github.com/eterline/ipcsv2base/internal/model"
)

// ipMetadata stores IP-related metadata including ASN, organization, and geolocation codes.

// RegistryIP represents a lookup registry for IP metadata.
type RegistryIP struct {
	reg          *IPContainerSet[networkMeta]
	countryTable []countryData
	asTable      []asData
}

// NewRegistryIP constructs a new RegistryIP by reading ASN and country CSV files.
// ver specifies the IP version filter (IPv4, IPv6, or both).
func NewRegistryIP(ctx context.Context, countryCSV, asnCSV string, ver IPVersion) (*RegistryIP, error) {

	countryTable := newUniquePrefixTable[countryData](0)
	astable := newUniquePrefixTable[asData](0)

	if err := csvForEach(
		countryCSV, 4, ver,
		func(network netip.Prefix, fields []string) error {

			countryTable.Add(network, countryData{
				ContinentCode: fields[0],
				CountryCode:   fields[1],
				CountryName:   fields[2],
			})

			return nil
		},
	); err != nil {
		return nil, fmt.Errorf("failed to read CSV file %s: %w", countryCSV, err)
	}

	if err := csvForEach(
		asnCSV, 6, ver,
		func(network netip.Prefix, fields []string) error {

			asn, err := strconv.ParseInt(fields[0], 0, 32)
			if err != nil {
				return err
			}

			astable.Add(network, asData{
				Number:      int32(asn),
				CountryCode: fields[1],
				Name:        fields[2],
				Org:         fields[3],
				Domain:      fields[4],
			})

			return nil
		},
	); err != nil {
		return nil, fmt.Errorf("failed to read CSV file %s: %w", asnCSV, err)
	}

	set := NewIPContainerSet[networkMeta](1 << 22)
	// set.Prepare()

	return &RegistryIP{reg: set}, nil
}

// Size returns the number of IP prefixes in the registry.
func (base *RegistryIP) Size() int {
	return base.reg.Size()
}

// LookupIP returns metadata for a given IP address.
func (base *RegistryIP) LookupIP(ctx context.Context, addr netip.Addr) (*model.IPMetadata, error) {
	pfx, meta, ok := base.reg.Get(addr)
	if !ok {
		return nil, errors.New("failed lookup")
	}

	data := &model.IPMetadata{
		Type:    model.NetworkGlobal,
		Network: pfx,
	}

	if id, ok := meta.getCountryID(); ok {
		c := base.countryTable[id]
		data.Geo = model.IPGeo{
			ContinentCode: model.GeoCode(c.ContinentCode),
			CountryCode:   model.GeoCode(c.CountryCode),
			CountryName:   c.CountryName,
		}
	}

	if id, ok := meta.getAsID(); ok {
		c := base.asTable[id]
		data.ASN = model.IPAS{
			ASN:         c.Number,
			CountryCode: model.GeoCode(c.CountryCode),
			Name:        c.Name,
			Org:         c.Org,
			Domain:      c.Domain,
		}
	}

	return data, nil
}

// ===============================

type countryData struct {
	ContinentCode string
	CountryCode   string
	CountryName   string
}

type asData struct {
	Number      int32
	CountryCode string
	Name        string
	Org         string
	Domain      string
}

type networkMeta struct {
	countryID uint32
	asID      uint32
}

func (m *networkMeta) setCountryID(id int) {
	if m == nil {
		panic("networkMeta is nil")
	}
	m.countryID = uint32(id)

}

func (m *networkMeta) setAsID(id int) {
	if m == nil {
		panic("networkMeta is nil")
	}
	m.asID = uint32(id)
}

func (m networkMeta) getCountryID() (id uint32, ok bool) {
	if m.countryID > 0 {
		return m.countryID, true
	}
	return 0, false
}

func (m networkMeta) getAsID() (id uint32, ok bool) {
	if m.asID > 0 {
		return m.asID, true
	}
	return 0, false
}

func (m networkMeta) valid() bool {
	return (m.countryID > 0) && (m.asID > 0)
}
