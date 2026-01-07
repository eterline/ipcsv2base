// Copyright (c) 2026 EterLine (Andrew)
// This file is part of fstmon.
// Licensed under the MIT License. See the LICENSE file for details.
package server

import "crypto/tls"

var (
	// tlsCiphers – preferred TLS cipher suites.
	tlsCiphers = []uint16{
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_RSA_WITH_AES_256_CBC_SHA,
	}

	// tlsCurves – preferred elliptic curves.
	tlsCurves = []tls.CurveID{
		tls.CurveP521,
		tls.CurveP384,
		tls.CurveP256,
	}

	// singleton TLS config
	defaultTlsConfig *tls.Config
)

/*
NewServerTlsConfig – creates or returns singleton TLS config for servers.
PreferServerCipherSuites is always true. Min TLS 1.2, Max TLS 1.3.
*/
func NewServerTlsConfig() *tls.Config {
	if defaultTlsConfig == nil {
		defaultTlsConfig = &tls.Config{
			PreferServerCipherSuites: true,
			CurvePreferences:         tlsCurves,
			CipherSuites:             tlsCiphers,
			MinVersion:               tls.VersionTLS12,
			MaxVersion:               tls.VersionTLS13,
		}
	}
	return defaultTlsConfig
}
