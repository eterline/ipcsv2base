package ipbase

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"strconv"
	"unique"

	"github.com/eterline/ipcsv2base/internal/model"
	mmaprc "github.com/eterline/ipcsv2base/pkg/mmapread"
)

// ipMetadata stores IP-related metadata including ASN, organization, and geolocation codes.
type ipMetadata struct {
	ASN           int32
	Name          unique.Handle[string]
	Org           unique.Handle[string]
	Domain        unique.Handle[string]
	ContinentCode code2bytes
	CountryCode   code2bytes
	CountryName   unique.Handle[string]
}

// RegistryIP represents a lookup registry for IP metadata.
type RegistryIP struct {
	reg *IPContainerSet[ipMetadata]
}

// NewRegistryIP constructs a new RegistryIP by reading ASN and country CSV files.
// ver specifies the IP version filter (IPv4, IPv6, or both).
func NewRegistryIP(ctx context.Context, countryCSV, asnCSV string, ver IPVersion) (*RegistryIP, error) {
	intern := newInterner()

	if err := readCountryCSV(countryCSV, intern, ver); err != nil {
		return nil, fmt.Errorf("failed to read CSV file %s: %w", countryCSV, err)
	}

	if err := readAsnCSV(asnCSV, intern, ver); err != nil {
		return nil, fmt.Errorf("failed to read CSV file %s: %w", asnCSV, err)
	}

	set := NewIPContainerSet[ipMetadata](1 << 22)
	for pfx, meta := range intern.meta {
		set.AddPrefix(pfx, meta)
	}

	intern.Clear()
	set.Prepare()

	return &RegistryIP{reg: set}, nil
}

// Size returns the number of IP prefixes in the registry.
func (base *RegistryIP) Size() int {
	return base.reg.Size()
}

// LookupIP returns metadata for a given IP address.
func (base *RegistryIP) LookupIP(ctx context.Context, addr netip.Addr) (*model.IPMetadata, error) {
	pfx, m, ok := base.reg.Get(addr)
	if !ok {
		return nil, errors.New("failed lookup")
	}

	data := &model.IPMetadata{
		Type:    model.NetworkGlobal,
		Network: pfx,
		Geo: model.IPGeo{
			ContinentCode: model.GeoCode(m.ContinentCode.String()),
			CountryCode:   model.GeoCode(m.CountryCode.String()),
			CountryName:   m.CountryName.Value(),
		},
		ASN: model.IPAS{
			ASN:         m.ASN,
			CountryCode: model.GeoCode(m.CountryCode.String()),
			Name:        m.Name.Value(),
			Org:         m.Org.Value(),
			Domain:      m.Domain.Value(),
		},
	}

	return data, nil
}

// readCountryCSV reads country CSV file and populates the interner with metadata.
func readCountryCSV(file string, intern *interner, ver IPVersion) error {
	f, err := mmaprc.OpenMMapReadCloser(file)
	if err != nil {
		return err
	}
	defer f.Close()

	countryRD := csv.NewReader(f)
	countryRD.FieldsPerRecord = 4

	if _, err := countryRD.Read(); err != nil { // skip header
		return err
	}

	for {
		recs, err := countryRD.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		pfx, err := netip.ParsePrefix(recs[0])
		if err != nil || !ver.validate(pfx.Addr()) {
			continue
		}

		meta := ipMetadata{
			ASN:           0,
			Name:          unique.Make(""),
			Org:           unique.Make(""),
			Domain:        unique.Make(""),
			ContinentCode: makecode2bytes(recs[1]),
			CountryCode:   makecode2bytes(recs[2]),
			CountryName:   unique.Make(recs[3]),
		}

		intern.c2c[meta.CountryCode] = meta.ContinentCode
		intern.meta[pfx] = meta
	}

	return nil
}

// readAsnCSV reads ASN CSV file and updates the interner with ASN and organization info.
func readAsnCSV(file string, intern *interner, ver IPVersion) error {
	f, err := mmaprc.OpenMMapReadCloser(file)
	if err != nil {
		return err
	}
	defer f.Close()

	asnRD := csv.NewReader(f)
	asnRD.FieldsPerRecord = 6

	if _, err := asnRD.Read(); err != nil { // skip header
		return err
	}

	for {
		recs, err := asnRD.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		pfx, err := netip.ParsePrefix(recs[0])
		if err != nil || !ver.validate(pfx.Addr()) {
			continue
		}

		asn, err := strconv.ParseInt(recs[1], 0, 32)
		if err != nil {
			continue
		}

		meta, ok := intern.meta[pfx]
		if !ok {
			country := makecode2bytes(recs[2])
			meta.ContinentCode = intern.c2c[country]
			meta.CountryCode = country
			meta.CountryName = unique.Make("")
		}

		meta.ASN = int32(asn)
		meta.Name = unique.Make(recs[3])
		meta.Org = unique.Make(recs[4])
		meta.Domain = unique.Make(recs[5])

		intern.meta[pfx] = meta
	}

	return nil
}
