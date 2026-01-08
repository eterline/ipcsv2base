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

		if err := do(pfx, recs[1:]); err != nil {
			return err
		}
	}

	return nil
}

// ==========================

const maxMetaIds = 1 << 24

type uniquePrefixTable[T comparable] struct {
	index map[T]uint32
	data  []T
	nets  [][]netip.Prefix
}

func newUniquePrefixTable[T comparable](capHint int) *uniquePrefixTable[T] {
	return &uniquePrefixTable[T]{
		index: make(map[T]uint32, capHint),
		data:  make([]T, 0, capHint),
		nets:  make([][]netip.Prefix, 0, capHint),
	}
}

func (t *uniquePrefixTable[T]) Add(pfx netip.Prefix, c T) {
	if id, ok := t.index[c]; ok {
		t.nets[id-1] = append(t.nets[id-1], pfx)
		return
	}

	id := uint32(len(t.data) + 1)
	if id > maxMetaIds {
		panic("too many indexes")
	}

	t.index[c] = id
	t.data = append(t.data, c)
	t.nets = append(t.nets, []netip.Prefix{pfx})
}

func (t *uniquePrefixTable[T]) Table() []T {
	return t.data
}

func (t *uniquePrefixTable[T]) Prefixes(id uint32) []netip.Prefix {
	return t.nets[id-1]
}

func (t *uniquePrefixTable[T]) TableForEach(f func(id uint32, prefixes []netip.Prefix, data T)) {
	for _, id := range t.index {
		f(id, t.Prefixes(id), t.data[id-1])
	}
}

func (t *uniquePrefixTable[T]) Clear() {
	clrSliceOfSlices(&t.nets)
	clrMap(&t.index)
}

// ==============

func clrMap[K comparable, V any](mPtr *map[K]V) {
	m := *mPtr
	for key := range m {
		delete(m, key)
	}
	*mPtr = nil
}

func clrSliceOfSlices[V any](sPtr *[][]V) {
	s := *sPtr
	for i := range s {
		s[i] = nil
	}
	*sPtr = nil
}
