package ipbase

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/netip"
	"runtime"
	"strconv"
	"unique"

	"github.com/eterline/ipcsv2base/internal/model"
	mmaprc "github.com/eterline/ipcsv2base/pkg/mmapread"
)

type code2bytes [2]byte

var emptyCode = code2bytes{0, 0}

func makecode2bytes(s string) code2bytes {
	if len(s) >= 2 {
		return [2]byte{s[0], s[1]}
	}
	return emptyCode
}

func (c2b code2bytes) String() string {
	if c2b == emptyCode {
		return ""
	}
	return string(c2b[:])
}

type ipMetadata struct {
	ASN           int32
	Name          unique.Handle[string]
	Org           unique.Handle[string]
	Domain        unique.Handle[string]
	ContinentCode code2bytes
	CountryCode   code2bytes
	CountryName   unique.Handle[string]
}

type metaInterner map[netip.Prefix]ipMetadata

func (mi metaInterner) Clear() {
	for key := range mi {
		delete(mi, key)
	}
}

type RegistryIP struct {
	reg *IPContainerSet[ipMetadata]
}

func NewRegistryIP(ctx context.Context, countryCSV, asnCSV string) (*RegistryIP, error) {
	internMeta := metaInterner{}

	if err := readCountryCSV(countryCSV, &internMeta); err != nil {
		return nil, fmt.Errorf("failed to read CSV file %s: %w", countryCSV, err)
	}

	if err := readAsnCSV(asnCSV, &internMeta); err != nil {
		return nil, fmt.Errorf("failed to read CSV file %s: %w", asnCSV, err)
	}

	set := NewIPContainerSet[ipMetadata](1 << 22)

	for pfx, meta := range internMeta {
		set.AddPrefix(pfx, meta)
	}

	internMeta.Clear()
	set.Prepare()
	runtime.GC()

	return &RegistryIP{reg: set}, nil
}

func (base *RegistryIP) Size() int {
	return base.reg.Size()
}

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

func readCountryCSV(file string, intern *metaInterner) error {
	f, err := mmaprc.OpenMMapReadCloser(file)
	if err != nil {
		return err
	}
	defer f.Close()

	countryRD := csv.NewReader(f)
	countryRD.FieldsPerRecord = 4

	if _, err := countryRD.Read(); err != nil {
		return err
	}

	internMeta := *intern

	for {
		recs, err := countryRD.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		pfx, err := netip.ParsePrefix(recs[0])
		if err != nil {
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

		internMeta[pfx] = meta
	}

	return nil
}

func readAsnCSV(file string, intern *metaInterner) error {
	f, err := mmaprc.OpenMMapReadCloser(file)
	if err != nil {
		return err
	}
	defer f.Close()

	countryRD := csv.NewReader(f)
	countryRD.FieldsPerRecord = 6

	if _, err := countryRD.Read(); err != nil {
		return err
	}

	internMeta := *intern

	for {
		recs, err := countryRD.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		pfx, err := netip.ParsePrefix(recs[0])
		if err != nil {
			continue
		}

		asn, err := strconv.ParseInt(recs[1], 0, 32)
		if err != nil {
			continue
		}

		meta, ok := internMeta[pfx]
		if !ok {
			meta.ContinentCode = emptyCode
			meta.CountryCode = makecode2bytes(recs[2])
			meta.CountryName = unique.Make("")
		}

		meta.ASN = int32(asn)
		meta.Name = unique.Make(recs[3])
		meta.Org = unique.Make(recs[4])
		meta.Domain = unique.Make(recs[5])

		internMeta[pfx] = meta
	}

	return nil
}
