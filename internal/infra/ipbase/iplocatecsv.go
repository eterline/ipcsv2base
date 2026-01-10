package ipbase

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"strconv"

	"github.com/eterline/ipcsv2base/internal/model"
	"github.com/eterline/ipcsv2base/pkg/ipsetdata"
)

// RegistryIP represents a lookup registry for IP metadata.
type RegistryIP struct {
	reg          *ipsetdata.IPContainerSet[networkMeta]
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

	netMap := map[netip.Prefix]networkMeta{}
	cIDByC := map[string]uint32{}

	countryTable.TableForEach(func(id uint32, prefixes []netip.Prefix, data countryData) {
		for _, pfx := range prefixes {
			netMap[pfx] = networkMeta{countryID: id}
			if _, ok := cIDByC[data.CountryCode]; !ok {
				cIDByC[data.CountryCode] = id
			}
		}
	})
	countryTable.Clear()

	astable.TableForEach(func(id uint32, prefixes []netip.Prefix, data asData) {
		for _, pfx := range prefixes {
			meta, ok := netMap[pfx]
			if !ok {
				meta = networkMeta{countryID: cIDByC[data.CountryCode]}
			}
			meta.setAsID(id)
			netMap[pfx] = meta
		}
	})

	set := ipsetdata.NewIPContainerSet[networkMeta](1 << 22)
	for pfx, meta := range netMap {
		set.AddPrefix(pfx, meta)
	}
	astable.Clear()

	clrMap(&netMap)
	clrMap(&cIDByC)

	set.Prepare()

	reg := &RegistryIP{
		reg:          set,
		countryTable: countryTable.Table(),
		asTable:      astable.Table(),
	}

	return reg, nil
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

	if idx, ok := meta.getCountryIdxID(); ok {
		c := base.countryTable[idx]
		data.Geo = model.IPGeo{
			ContinentCode: model.GeoCode(c.ContinentCode),
			CountryCode:   model.GeoCode(c.CountryCode),
			CountryName:   c.CountryName,
		}
	}

	if idx, ok := meta.getAsIdxID(); ok {
		c := base.asTable[idx]
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

func (m *networkMeta) setAsID(id uint32) {
	if m == nil {
		panic("networkMeta is nil")
	}
	m.asID = uint32(id)
}

func (m networkMeta) getCountryIdxID() (id uint32, ok bool) {
	if m.countryID > 0 {
		return m.countryID - 1, true
	}
	return 0, false
}

func (m networkMeta) getAsIdxID() (id uint32, ok bool) {
	if m.asID > 0 {
		return m.asID - 1, true
	}
	return 0, false
}
