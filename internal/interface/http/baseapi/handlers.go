package baseapi

import (
	"context"
	"errors"
	"net/http"
	"net/netip"
	"strings"
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
	log      model.Logger
	ipLookup *security.IpExtractor
}

// NewBaseAPIHandlerGroup - Creates a new API handler group for IP base lookups.
func NewBaseAPIHandlerGroup(log model.Logger, l Lookuper, lookupHeadersIp bool) *BaseAPIHandlerGroup {
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
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

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
	log := h.log.With(model.FieldStringer("ip", addr))

	// Perform lookup
	meta, err := h.lookup.LookupIP(ctx, addr)
	if err != nil {
		log.Error("ip lookup failed", model.FieldError(err))
		api.NewResponse().
			SetCode(http.StatusInternalServerError).
			SetMessage("lookup failed").
			Write(w)
		return
	}

	dto := domain2IPMetadataDTO(meta, time.Since(startAt), addr)
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
func (h *BaseAPIHandlerGroup) LookupSubnetHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	rawNet := chi.URLParam(r, "net")
	if strings.Contains(rawNet, " ") {
		cleanNet := strings.ReplaceAll(rawNet, " ", "")
		newPath := strings.Replace(r.URL.Path, rawNet, cleanNet, 1)
		http.Redirect(w, r, newPath, http.StatusFound)
		return
	}

	log := h.log.With(model.FieldString("prefix", rawNet))

	startAt := time.Now()

	pfx, err := netip.ParsePrefix(rawNet)
	if err != nil {
		log.Debug("invalid network prefix", model.FieldError(err))
		api.NewResponse().
			SetCode(http.StatusInternalServerError).
			SetMessage("invalid subnet").
			AddStringError(err.Error()).
			Write(w)
		return
	}

	meta, err := h.lookup.LookupPrefix(ctx, pfx)
	if err != nil {
		log.Error("prefix lookup failed", model.FieldError(err))
		api.NewResponse().
			SetCode(http.StatusInternalServerError).
			SetMessage("lookup failed").
			Write(w)
		return
	}

	dto := domain2IPMetadataDTO(meta, time.Since(startAt), pfx.Addr())
	api.NewResponse().
		SetCode(http.StatusOK).
		WrapData(dto).
		Write(w)
}

func (h *BaseAPIHandlerGroup) AvailableTypes() func(w http.ResponseWriter, r *http.Request) {
	types := map[string]string{
		model.NetworkGlobal.String():   "Global public IP address, reachable from the Internet (RFC 791). Example: 8.8.8.8",
		model.NetworkPrivate.String():  "Private IP address, used inside local networks, not routed on the Internet (RFC 1918). Examples: 192.168.0.0/16, 10.0.0.0/8",
		model.NetworkLoopback.String(): "Loopback address for self-communication, used for testing local services (RFC 1122). Examples: 127.0.0.1, ::1",
		model.NetworkTest.String():     "Test/documentation addresses reserved for examples/testing, not routable in production (RFC 5737). Examples: 192.0.2.0/24, 198.51.100.0/24",
	}

	return func(w http.ResponseWriter, r *http.Request) {
		api.NewResponse().
			SetCode(http.StatusOK).
			WrapData(types).
			Write(w)
	}
}
