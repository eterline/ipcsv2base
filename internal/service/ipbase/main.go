package ipbase

import (
	"context"
	"errors"
	"log/slog"
	"net/netip"

	"github.com/eterline/ipcsv2base/internal/model"
)

/*
MetaLookuper - Interface for resolving IP metadata from a primary data source.
Implementations are expected to perform a full lookup (e.g. radix tree, mmap base).
*/
type MetaLookuper interface {
	LookupIP(ctx context.Context, addr netip.Addr) (*model.IPMetadata, error)
}

/*
MetaCache - Interface for IP metadata caching layer.
Used to speed up repeated lookups and reduce pressure on the main lookup source.
*/
type MetaCache interface {
	LookupIP(ctx context.Context, addr netip.Addr) (*model.IPMetadata, bool)
	SaveIP(addr netip.Addr, meta *model.IPMetadata)
}

/*
IPBaseService - Core service for IP metadata lookup with cache and logging support.
It classifies network type, handles cache hits, and delegates lookups to MetaLookuper.
*/
type IPBaseService struct {
	lookup MetaLookuper
	cache  MetaCache
	log    *slog.Logger
}

/*
NewIPBaseService - Constructs a new IPBaseService instance.

Parameters:
  - log: structured logger instance
  - l: primary metadata lookuper
  - c: cache implementation
*/
func NewIPBaseService(
	log *slog.Logger,
	l MetaLookuper,
	c MetaCache,
) *IPBaseService {
	return &IPBaseService{
		lookup: l,
		cache:  c,
		log:    log,
	}
}

/*
LookupIP - Performs metadata lookup for a single IP address.

Lookup flow:
 1. Detect network type (global / private / test).
 2. Reject unknown network areas.
 3. Return minimal metadata for private and test networks.
 4. Try cache lookup.
 5. Fallback to primary lookuper.
 6. Save result to cache asynchronously.
*/
func (b *IPBaseService) LookupIP(
	ctx context.Context,
	addr netip.Addr,
) (*model.IPMetadata, error) {

	nt, pfx := NetworkTypeFromAddrWithSubnet(addr)

	log := b.log.With(
		"ip", addr.String(),
		"network_type", nt.String(),
	)

	switch nt {
	case model.NetworkUnknown:
		log.Error("lookup aborted: unknown network type")
		return nil, errors.New("unknown network area")

	case model.NetworkPrivate,
		model.NetworkTest,
		model.NetworkLoopback:
		log.Debug("lookup skipped: non-global network")
		return &model.IPMetadata{Type: nt, Network: pfx}, nil
	}

	// Cache lookup
	if meta, ok := b.cache.LookupIP(ctx, addr); ok {
		log.Info("cache hit")
		return meta, nil
	}

	// Primary lookup
	meta, err := b.lookup.LookupIP(ctx, addr)
	if err != nil {
		log.Error("lookup failed", "error", err)
		return nil, err
	}

	// Async cache save
	go func() {
		log.Debug("saving result to cache")
		b.cache.SaveIP(addr, meta)
	}()

	log.Debug("lookup finished")
	return meta, nil
}

/*
LookupPrefix - Performs metadata lookup for a network prefix.
Currently resolves metadata based on the prefix address.
*/
func (b *IPBaseService) LookupPrefix(
	ctx context.Context,
	pfx netip.Prefix,
) (*model.IPMetadata, error) {
	return b.LookupIP(ctx, pfx.Addr())
}
