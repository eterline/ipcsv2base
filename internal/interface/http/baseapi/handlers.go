package baseapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/netip"
	"time"

	"github.com/eterline/ipcsv2base/internal/interface/http/api"
	"github.com/eterline/ipcsv2base/internal/interface/security"
	"github.com/eterline/ipcsv2base/internal/model"
	"github.com/go-chi/chi/v5"
)

type Lookuper interface {
	LookupIP(context.Context, netip.Addr) (*model.IPMetadata, error)
	LookupPrefix(context.Context, netip.Prefix) (*model.IPMetadata, error)
}

type BaseAPIHandlerGroup struct {
	lookup   Lookuper
	log      *slog.Logger
	ipLookup *security.IpExtractor
}

// NewBaseAPIHandlerGroup - Creates a new API handler group for IP base lookups.
func NewBaseAPIHandlerGroup(
	log *slog.Logger,
	l Lookuper,
	lookupHeadersIp bool,
) *BaseAPIHandlerGroup {
	return &BaseAPIHandlerGroup{
		lookup:   l,
		log:      log,
		ipLookup: security.NewIpExtractor(lookupHeadersIp),
	}
}

// extractIP - Extracts IP from path parameter or falls back to request.
// Returns user-visible errors for parsing failures.
func (h *BaseAPIHandlerGroup) extractIP(r *http.Request) (netip.Addr, error) {
	rawIP := chi.URLParam(r, "ip")

	if rawIP != "" {
		addr, err := netip.ParseAddr(rawIP)
		if err != nil {
			return netip.Addr{}, errors.New("invalid ip address in path")
		}
		return addr, nil
	}

	// Fallback: extract IP from request headers or RemoteAddr
	if h.ipLookup != nil {
		addr, _, err := h.ipLookup.ExtractIP(r)
		if err != nil {
			return netip.Addr{}, errors.New("invalid ip address in request")
		}
		return addr, nil
	}

	return netip.Addr{}, errors.New("no ip provided")
}

// LookupIPHandler - Handles metadata lookup for a single IP address.
//
// Path parameters:
//   - ip: IPv4 or IPv6 address
//
// Parsing errors are returned to the client.
// Internal lookup errors are logged and hidden.
func (h *BaseAPIHandlerGroup) LookupIPHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	startAt := time.Now()

	// Extract IP either from path or request
	addr, err := h.extractIP(r)
	if err != nil {
		api.NewResponse().
			SetCode(http.StatusBadRequest).
			SetMessage(err.Error()).
			Write(w)
		return
	}

	// Prepare structured log
	log := h.log.With("ip", addr.String())

	// Perform lookup
	meta, err := h.lookup.LookupIP(ctx, addr)
	if err != nil {
		log.Error("ip lookup failed", "error", err)
		api.NewResponse().
			SetCode(http.StatusInternalServerError).
			SetMessage("lookup failed").
			Write(w)
		return
	}

	dto := domain2IPMetadataDTO(meta, time.Since(startAt))
	api.NewResponse().
		SetCode(http.StatusOK).
		WrapData(dto).
		Write(w)
}

// LookupSubnetHandler - Handles metadata lookup for a network prefix.
//
// Path parameters:
//   - net: network prefix in CIDR notation
//
// Parsing errors are returned to the client.
// Internal lookup errors are logged and hidden.
func (h *BaseAPIHandlerGroup) LookupSubnetHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	ctx := r.Context()
	rawNet := chi.URLParam(r, "net")
	log := h.log.With("prefix", rawNet)

	startAt := time.Now()

	pfx, err := netip.ParsePrefix(rawNet)
	if err != nil {
		log.Debug("invalid network prefix", "error", err)
		api.NewResponse().
			SetCode(http.StatusInternalServerError).
			SetMessage("invalid subnet").
			AddStringError(err.Error()).
			Write(w)
		return
	}

	meta, err := h.lookup.LookupPrefix(ctx, pfx)
	if err != nil {
		log.Error("prefix lookup failed", "error", err)
		api.NewResponse().
			SetCode(http.StatusInternalServerError).
			SetMessage("lookup failed").
			Write(w)
		return
	}

	dto := domain2IPMetadataDTO(meta, time.Since(startAt))
	api.NewResponse().
		SetCode(http.StatusOK).
		WrapData(dto).
		Write(w)
}
