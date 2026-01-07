package ipbase

import (
	"context"
	"net/netip"

	"github.com/eterline/ipcsv2base/internal/model"
)

type IPbaseCacheMock struct{}

func (mock *IPbaseCacheMock) LookupIP(ctx context.Context, addr netip.Addr) (*model.IPMetadata, bool) {
	return nil, false
}

func (mock *IPbaseCacheMock) SaveIP(addr netip.Addr, meta *model.IPMetadata) {}

type IPbaseMock struct{}

func (mock *IPbaseMock) LookupIP(ctx context.Context, addr netip.Addr) (*model.IPMetadata, error) {
	return &model.IPMetadata{
		Type: model.NetworkGlobal,
	}, nil
}
