package ipsetdata

import (
	"net/netip"
	"sort"

	"go4.org/netipx"
)

type container[T comparable] struct {
	rng  rangeUint128t
	data T
}

type IPContainerSet[T comparable] struct {
	set []container[T]
}

func NewIPContainerSet[T comparable](prep int) *IPContainerSet[T] {
	return &IPContainerSet[T]{
		set: make([]container[T], 0, prep),
	}
}

func (cset *IPContainerSet[T]) AddIPRange(rng netipx.IPRange, value T) {
	cset.set = append(cset.set, container[T]{
		rng: rangeUint128t{
			start: Addr2Uint128t(rng.From()),
			end:   Addr2Uint128t(rng.To()),
		},
		data: value,
	})
}

func (cset *IPContainerSet[T]) AddPrefix(pfx netip.Prefix, value T) {
	cset.AddIPRange(netipx.RangeOfPrefix(pfx), value)
}

func (cset *IPContainerSet[T]) Prepare() {
	sort.Slice(cset.set, func(i, j int) bool {
		return cset.set[i].rng.start.Compare(cset.set[j].rng.start) < 0
	})

	cset.set = cset.set[:len(cset.set):len(cset.set)]
}

func (cset *IPContainerSet[T]) Get(ip netip.Addr) (pfx netip.Prefix, data T, ok bool) {
	ipvec := Addr2Uint128t(ip)

	i := sort.Search(len(cset.set), func(i int) bool {
		return !cset.set[i].rng.start.Less(ipvec)
	})

	if i < len(cset.set) && cset.set[i].rng.Contains(ipvec) {
		pfx, ok := cset.set[i].rng.ToIPRange().Prefix()
		if ok {
			return pfx, cset.set[i].data, true
		}
	}

	if i > 0 && cset.set[i-1].rng.Contains(ipvec) {
		pfx, ok := cset.set[i-1].rng.ToIPRange().Prefix()
		if ok {
			return pfx, cset.set[i-1].data, true
		}
	}

	var dflt T
	return netip.Prefix{}, dflt, false
}

func (cset *IPContainerSet[T]) Size() int {
	return len(cset.set)
}
