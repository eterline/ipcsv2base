package ipsetdata

import (
	"fmt"
	"net/netip"
	"sort"

	"go4.org/netipx"
)

// container - Internal structure that binds an IP range represented as uint128 values with associated data.
type container[T comparable] struct {
	rng  rangeUint128t
	data T
}

/*
IPContainerSet - Sorted set of IP ranges with associated data.

	Lookup is performed using binary search.
*/
type IPContainerSet[T comparable] struct {
	set []container[T]
}

// NewIPContainerSet - Creates a new IPContainerSet with preallocated capacity.
func NewIPContainerSet[T comparable](prep int) *IPContainerSet[T] {
	return &IPContainerSet[T]{
		set: make([]container[T], 0, prep),
	}
}

/*
AddIPRange - Adds an IP range with associated value to the set.

	The range is stored internally as uint128 boundaries.
*/
func (cset *IPContainerSet[T]) AddIPRange(rng netipx.IPRange, value T) {
	cset.set = append(cset.set, container[T]{
		rng: rangeUint128t{
			start: Addr2Uint128t(rng.From()),
			end:   Addr2Uint128t(rng.To()),
		},
		data: value,
	})
}

// AddStartEnd - Adds a CIDR prefix with associated value to the set.
func (cset *IPContainerSet[T]) AddStartEnd(start, end netip.Addr, value T) {
	cset.set = append(cset.set, container[T]{
		rng: rangeUint128t{
			start: Addr2Uint128t(start),
			end:   Addr2Uint128t(end),
		},
		data: value,
	})
}

// AddStartEndStrings - Adds a CIDR prefix with associated value to the set.
func (cset *IPContainerSet[T]) AddStartEndStrings(start, end string, value T) error {
	startAddr, err := netip.ParseAddr(start)
	if err != nil {
		return fmt.Errorf("invalid start range addr: %w", err)
	}

	endAddr, err := netip.ParseAddr(end)
	if err != nil {
		return fmt.Errorf("invalid end range addr: %w", err)
	}

	cset.AddStartEnd(startAddr, endAddr, value)
	return nil
}

// AddPrefix - Adds a CIDR prefix with associated value to the set.
func (cset *IPContainerSet[T]) AddPrefix(pfx netip.Prefix, value T) {
	cset.AddIPRange(netipx.RangeOfPrefix(pfx), value)
}

/*
Prepare - Sorts internal ranges by start address.

	Must be called before any Get() calls.
*/
func (cset *IPContainerSet[T]) Prepare() {
	sort.Slice(cset.set, func(i, j int) bool {
		return cset.set[i].rng.start.Less(cset.set[j].rng.start)
	})

	// Force capacity to length to prevent accidental reallocation
	cset.set = cset.set[:len(cset.set):len(cset.set)]
}

/*
Get - Finds the IP range containing the given address.

	Returns the matching prefix, associated data and true on success.
*/
func (cset *IPContainerSet[T]) Get(ip netip.Addr) (pfx netip.Prefix, data T, ok bool) {
	ipvec := Addr2Uint128t(ip)

	i := sort.Search(len(cset.set), func(i int) bool {
		return !cset.set[i].rng.start.Less(ipvec)
	})

	// Check exact index
	if i < len(cset.set) && cset.set[i].rng.Contains(ipvec) {
		if p, ok := cset.set[i].rng.ToIPRange().Prefix(); ok {
			return p, cset.set[i].data, true
		}
	}

	// Check previous range
	if i > 0 && cset.set[i-1].rng.Contains(ipvec) {
		if p, ok := cset.set[i-1].rng.ToIPRange().Prefix(); ok {
			return p, cset.set[i-1].data, true
		}
	}

	var zero T
	return netip.Prefix{}, zero, false
}

// Size - Returns number of stored ranges.
func (cset *IPContainerSet[T]) Size() int {
	return len(cset.set)
}
