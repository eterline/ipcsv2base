package ipbase

// import (
// 	"bufio"
// 	"bytes"
// 	"context"
// 	"errors"
// 	"fmt"
// 	"io"
// 	"net/netip"
// 	"unsafe"

// 	"github.com/eterline/ipcsv2base/internal/model"
// 	mmaprc "github.com/eterline/ipcsv2base/pkg/mmapread"
// )

// // RegistryContryIP represents a lookup registry for IP contry.
// type RegistryContryOnlyIP struct {
// 	reg *IPContainerSet[code2bytes]
// }

// // NewRegistryContryOnlyIP constructs a new RegistryContryIP by reading ASN and country CSV files.
// // ver specifies the IP version filter (IPv4, IPv6, or both).
// func NewRegistryContryOnlyIP(ctx context.Context, countryCSV string, ver IPVersion) (*RegistryContryOnlyIP, error) {
// 	set := NewIPContainerSet[code2bytes](1 << 22)

// 	err := iterateCountryCSV(ctx, countryCSV, ver,
// 		func(pfx netip.Prefix, rec string) {
// 			set.AddPrefix(pfx, makecode2bytes(rec))
// 		},
// 	)
// 	if err != nil {
// 		return nil, err
// 	}

// 	set.Prepare()

// 	return &RegistryContryOnlyIP{reg: set}, nil
// }

// // Size returns the number of IP prefixes in the registry.
// func (base *RegistryContryOnlyIP) Size() int {
// 	return base.reg.Size()
// }

// // LookupIP returns metadata for a given IP address.
// func (base *RegistryContryOnlyIP) LookupIP(ctx context.Context, addr netip.Addr) (*model.IPMetadata, error) {
// 	pfx, m, ok := base.reg.Get(addr)
// 	if !ok {
// 		return nil, errors.New("failed lookup")
// 	}

// 	data := &model.IPMetadata{
// 		Type:    model.NetworkGlobal,
// 		Network: pfx,
// 		Geo: model.IPGeo{
// 			CountryCode: model.GeoCode(m.String()),
// 		},
// 		ASN: model.IPAS{
// 			CountryCode: model.GeoCode(m.String()),
// 		},
// 	}

// 	return data, nil
// }

// func iterateCountryCSV(ctx context.Context, countryCSV string, ver IPVersion, fn func(pfx netip.Prefix, rec string)) error {
// 	f, err := mmaprc.OpenMMapReadCloser(countryCSV)
// 	if err != nil {
// 		return fmt.Errorf("failed to read CSV file %s: %w", countryCSV, err)
// 	}
// 	defer f.Close()

// 	br := bufio.NewReaderSize(f, 1<<12)

// 	// skip header
// 	if _, err := br.ReadSlice('\n'); err != nil {
// 		return fmt.Errorf("failed to read CSV header %s: %w", countryCSV, err)
// 	}

// 	var callFn = func(line []byte) error {
// 		s := bytes.Split(line, []byte{','})
// 		if len(s) < 2 {
// 			return errors.New("invalid line size")
// 		}

// 		pfx, err := netip.ParsePrefix(bytesToString(s[0]))
// 		if err != nil {
// 			return err
// 		}

// 		fn(pfx, bytesToString(s[1]))
// 		return nil
// 	}

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return ctx.Err()
// 		default:
// 		}

// 		line, err := br.ReadSlice('\n')
// 		if err != nil {
// 			if err == bufio.ErrBufferFull {
// 				return fmt.Errorf("line too long")
// 			}
// 			if err == io.EOF {
// 				if len(line) > 0 {
// 					return callFn(line)
// 				}
// 				return nil
// 			}
// 			return err
// 		}

// 		if err := callFn(line); err != nil {
// 			return fmt.Errorf("failed to read CSV header %s: %w", countryCSV, err)
// 		}
// 	}
// }

// func bytesToString(b []byte) string {
// 	return unsafe.String(unsafe.SliceData(b), len(b))
// }
