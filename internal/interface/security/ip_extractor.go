// Copyright (c) 2025 EterLine (Andrew)
// This file is part of fstmon.
// Licensed under the MIT License. See the LICENSE file for details.
package security

import (
	"net/http"
	"net/netip"
	"strings"
)

// IpExtractor – structure for extracting the client IP address from an HTTP request.
// If headers is true, ExtractIP will attempt to read the IP from request headers
// ('X-Real-IP', 'X-Forwarded-For', 'Forwarded') before falling back to RemoteAddr.
type IpExtractor struct {
	headers bool
}

// NewIpExtractor – constructor for IpExtractor.
// headers: if true, ExtractIP will try to extract the IP from headers.
func NewIpExtractor(headers bool) *IpExtractor {
	return &IpExtractor{
		headers: headers,
	}
}

// ExtractIP – extracts the IP address from the HTTP request.
// If headers are enabled, it checks headers first. Otherwise, it uses RemoteAddr.
// Returns netip.Addr and an error if parsing fails.
func (is *IpExtractor) ExtractIP(r *http.Request) (client netip.Addr, remote netip.AddrPort, err error) {
	rmtAp, _ := Remote(r)
	// if err != nil {
	// 	return netip.Addr{}, netip.AddrPort{}, fmt.Errorf("ip extarctor error: %w", err)
	// }

	if is.headers {
		if ip, ok := headers(r); ok {
			return ip, rmtAp, nil
		}
	}

	return rmtAp.Addr(), rmtAp, nil
}

// headers – tries to extract the IP from HTTP headers.
// Returns the IP and true if successfully parsed, otherwise false.
func headers(r *http.Request) (netip.Addr, bool) {
	if r.Header == nil {
		return netip.Addr{}, false
	}

	if ip, ok := ParseXRealIP(r.Header); ok {
		return ip, true
	}
	if ip, ok := ParseXForwardedFor(r.Header); ok {
		return ip, true
	}
	if ip, ok := ParseForwarded(r.Header); ok {
		return ip, true
	}
	return netip.Addr{}, false
}

// Remote – extracts the IP from r.RemoteAddr.
// Returns an error if RemoteAddr is invalid or cannot be parsed.
func Remote(r *http.Request) (netip.AddrPort, error) {
	addrPort, err := netip.ParseAddrPort(r.RemoteAddr)
	if err != nil {
		return netip.AddrPort{}, err
	}

	return addrPort, nil
}

// ParseXRealIP – parses the 'X-Real-IP' header (used by Nginx reverse proxy).
// Returns the parsed IP and true if successful.
func ParseXRealIP(h http.Header) (netip.Addr, bool) {
	value := h.Get("X-Real-IP")
	if value == "" {
		return netip.Addr{}, false
	}

	ipStr := strings.TrimSpace(value)
	ip, err := netip.ParseAddr(ipStr)
	if err != nil {
		return netip.Addr{}, false
	}
	return ip, true
}

// ParseXForwardedFor – parses the 'X-Forwarded-For' header.
// Returns the first IP in the list and true if successful.
func ParseXForwardedFor(h http.Header) (netip.Addr, bool) {
	value := h.Get("X-Forwarded-For")
	if value == "" {
		return netip.Addr{}, false
	}

	parts := strings.Split(value, ",")
	if len(parts) == 0 {
		return netip.Addr{}, false
	}

	ipStr := strings.TrimSpace(parts[0])
	ip, err := netip.ParseAddr(ipStr)
	if err != nil {
		return netip.Addr{}, false
	}
	return ip, true
}

// ParseForwarded – parses the 'Forwarded' header (RFC 7239).
// Looks for 'for=' directives and returns the first valid IP found.
func ParseForwarded(h http.Header) (netip.Addr, bool) {
	value := h.Get("Forwarded")
	if value == "" {
		return netip.Addr{}, false
	}

	entries := strings.Split(value, ",")
	for _, entry := range entries {
		parts := strings.Split(entry, ";")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if strings.HasPrefix(strings.ToLower(p), "for=") {
				val := strings.TrimPrefix(p, "for=")
				val = strings.Trim(val, "\"[]") // remove quotes/brackets
				ip, err := netip.ParseAddr(val)
				if err == nil {
					return ip, true
				}
			}
		}
	}
	return netip.Addr{}, false
}
