package ipbase

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"net/netip"

	"github.com/eterline/ipcsv2base/internal/model"
	"github.com/eterline/ipcsv2base/pkg/ipsetdata"
	mmaprc "github.com/eterline/ipcsv2base/pkg/mmapread"
	"github.com/eterline/ipcsv2base/pkg/toolkit"
)

type RegistryIPTSV struct {
	reg *ipsetdata.IPContainerSet[uint16]
}

func NewRegistryIPTSV(ctx context.Context, files ...string) (*RegistryIPTSV, error) {
	set := ipsetdata.NewIPContainerSet[uint16](1 << 22)

	for _, file := range files {
		err := func() error {
			src, err := mmaprc.OpenMMapReadCloser(file)
			if err != nil {
				return err
			}
			defer src.Close()

			r := bufio.NewReaderSize(src, 1<<20)

			for {
				line, err := r.ReadSlice('\n')
				if err == bufio.ErrBufferFull {
					continue
				}
				if err == io.EOF {
					if len(line) > 0 {
						if err := addToSet(set, line); err != nil {
							return err
						}
					}
					break
				}
				if err != nil {
					return err
				}

				if err := addToSet(set, line); err != nil {
					return err
				}
			}

			return nil
		}()

		if err != nil {
			return nil, err
		}
	}

	set.Prepare()
	return &RegistryIPTSV{reg: set}, nil
}

// Size returns the number of IP prefixes in the registry.
func (base *RegistryIPTSV) Size() int {
	return base.reg.Size()
}

// LookupIP returns metadata for a given IP address.
func (base *RegistryIPTSV) LookupIP(ctx context.Context, addr netip.Addr) (*model.IPMetadata, error) {
	pfx, code, ok := base.reg.Get(addr)
	if !ok {
		return nil, errors.New("failed lookup")
	}

	codeBytes := make([]byte, 2)
	toolkit.Uint16ToBytesLE(code, codeBytes)

	data := &model.IPMetadata{
		Type:    model.NetworkGlobal,
		Network: pfx,
		Geo:     model.IPGeo{CountryCode: model.GeoCode(codeBytes)},
	}

	return data, nil
}

func addToSet(set *ipsetdata.IPContainerSet[uint16], line []byte) error {
	rec := bytes.Split(line, []byte{'\t'})
	if len(rec) != 3 {
		return nil
	}

	if len(rec[2]) < 2 {
		return nil
	}

	return set.AddStartEndStrings(
		toolkit.BytesToString(rec[0]),
		toolkit.BytesToString(rec[1]),
		toolkit.BytesToUint16LE(rec[2]),
	)
}
