package ipcsv2base

import (
	"fmt"
)

// ===================

type NetworkCountryWriter struct {
	base *BaseCreator
}

func CreateNetworkCountryBase(file string) (*NetworkCountryWriter, error) {
	base, err := CreateBase(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create country base file: %w", err)
	}

	return &NetworkCountryWriter{base: base}, nil
}

func (ncw *NetworkCountryWriter) Write(record NetworkCountry) (s int, err error) {
	return ncw.base.Add(record)
}

func (ncw *NetworkCountryWriter) Close() error {
	return ncw.base.Close()
}

func (ncw *NetworkCountryWriter) Writes() int64 {
	return ncw.base.Writes()
}

func (ncw *NetworkCountryWriter) Size() int64 {
	return ncw.base.Size()
}

// ===================

type NetworkCountryReader struct {
	allowedCodes map[[2]byte]struct{}
	base         *BaseReader
	ver          NetworkVersion
}

func OpenNetworkCountryBase(path string, ver NetworkVersion, codes ...string) (*NetworkCountryReader, error) {
	base, err := OpenBase(path, NetworkCountry{})
	if err != nil {
		return nil, fmt.Errorf("failed to open country base file: %w", err)
	}

	return &NetworkCountryReader{
		allowedCodes: genCodesMap(codes),
		base:         base,
		ver:          ver,
	}, nil
}

func (r *NetworkCountryReader) Next() (NetworkCountry, error) {
	var record NetworkCountry

	for {
		err := r.base.Next(&record)
		if err != nil {
			return NetworkCountry{}, err
		}

		pfx := record.Prefix()

		if (r.ver == IPv4 && !pfx.Addr().Is4()) || (r.ver == IPv6 && !pfx.Addr().Is6()) {
			continue
		}

		if r.allowedCodes != nil {
			if _, ok := r.allowedCodes[record.counrtyCode()]; !ok {
				continue
			}
		}

		return record, nil
	}
}

func (ncw *NetworkCountryReader) Reset() error {
	return ncw.base.Reset()
}

func (ncw *NetworkCountryReader) Close() error {
	return ncw.base.Close()
}

func genCodesMap(codes []string) (m map[[2]byte]struct{}) {
	if len(codes) == 0 {
		return nil
	}

	m = make(map[[2]byte]struct{})

	for _, code := range codes {
		if len(code) != 2 {
			continue
		}

		m[[2]byte{code[0], code[1]}] = struct{}{}
	}

	return
}
