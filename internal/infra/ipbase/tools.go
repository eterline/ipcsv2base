package ipbase

import (
	"encoding/csv"
	"io"
	"net/netip"

	mmaprc "github.com/eterline/ipcsv2base/pkg/mmapread"
)

// IPVersion represents the IP address version type.
type IPVersion uint8

const (
	IPv4v6 IPVersion = iota // Both IPv4 and IPv6
	IPv4                    // IPv4 only
	IPv6                    // IPv6 only
)

// IPVersionStr converts a string representation ("v4" or "v6") to IPVersion.
func IPVersionStr(s string) IPVersion {
	switch s {
	case "v4":
		return IPv4
	case "v6":
		return IPv6
	default:
		return IPv4v6
	}
}

// validate checks whether the given IP address matches the version.
func (v IPVersion) validate(ip netip.Addr) bool {
	switch v {
	case IPv4v6:
		return true
	case IPv4:
		return ip.Is4() || ip.Is4In6()
	case IPv6:
		return ip.Is6()
	default:
		return false
	}
}

// ==========================

type csvEachFunc func(network netip.Prefix, fields []string) error

// readAsnCSV reads ASN CSV file and updates the interner with ASN and organization info.
func csvForEach(file string, fieldsCount int, verAllow IPVersion, do csvEachFunc) error {
	f, err := mmaprc.OpenMMapReadCloser(file)
	if err != nil {
		return err
	}
	defer f.Close()

	asnRD := csv.NewReader(f)
	asnRD.FieldsPerRecord = fieldsCount

	// skip header
	if _, err := asnRD.Read(); err != nil {
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
		if err != nil || !verAllow.validate(pfx.Addr()) {
			continue
		}

		if err := do(pfx, recs[0:]); err != nil {
			return err
		}
	}

	return nil
}

// ==========================

type uniquePrefixTable[T comparable] struct {
	index map[T]int
	data  []T
	nets  [][]netip.Prefix
}

func newUniquePrefixTable[T comparable](capHint int) *uniquePrefixTable[T] {
	return &uniquePrefixTable[T]{
		index: make(map[T]int, capHint),
		data:  make([]T, 0, capHint),
		nets:  make([][]netip.Prefix, 0, capHint),
	}
}

func (t *uniquePrefixTable[T]) Add(pfx netip.Prefix, c T) int {
	if id, ok := t.index[c]; ok {
		t.nets[id] = append(t.nets[id], pfx)
		return id
	}

	id := len(t.data)
	t.index[c] = id
	t.data = append(t.data, c)
	t.nets = append(t.nets, []netip.Prefix{pfx})

	return id
}

func (t *uniquePrefixTable[T]) Data(id int) T {
	return t.data[id]
}

func (t *uniquePrefixTable[T]) Prefixes(id int) []netip.Prefix {
	return t.nets[id]
}
