package ipbase

import (
	"net/netip"
	"sort"
	"unique"

	"go4.org/netipx"
)

type ContainerEntry[T comparable] struct {
	rng  netipx.IPRange
	data unique.Handle[T]
}

type IPContainerSet[T comparable] struct {
	set []ContainerEntry[T]
}

func NewIPContainerSet[T comparable](prep int) *IPContainerSet[T] {
	return &IPContainerSet[T]{
		set: make([]ContainerEntry[T], 0, prep),
	}
}

func (cset *IPContainerSet[T]) AddIPRange(rng netipx.IPRange, value T) {
	entry := ContainerEntry[T]{
		rng:  rng,
		data: unique.Make(value),
	}
	cset.set = append(cset.set, entry)
}

func (cset *IPContainerSet[T]) AddPrefix(pfx netip.Prefix, value T) {
	cset.AddIPRange(netipx.RangeOfPrefix(pfx), value)
}

func (cset *IPContainerSet[T]) Prepare() {
	sort.Slice(cset.set, func(i, j int) bool {
		return cset.set[i].rng.From().Compare(cset.set[j].rng.From()) < 0
	})

	cset.set = cset.set[:len(cset.set):len(cset.set)]
}

func (cset *IPContainerSet[T]) Get(ip netip.Addr) (pfx netip.Prefix, data T, ok bool) {

	i := sort.Search(len(cset.set), func(i int) bool {
		return !cset.set[i].rng.From().Less(ip)
	})

	if i < len(cset.set) && cset.set[i].rng.Contains(ip) {
		pfx, ok := cset.set[i].rng.Prefix()
		if ok {
			return pfx, cset.set[i].data.Value(), true
		}
	}

	if i > 0 && cset.set[i-1].rng.Contains(ip) {
		pfx, ok := cset.set[i].rng.Prefix()
		if ok {
			return pfx, cset.set[i-1].data.Value(), true
		}
	}

	var dflt T
	return netip.Prefix{}, dflt, false
}

func (cset *IPContainerSet[T]) Size() int {
	return len(cset.set)
}
