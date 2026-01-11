// Copyright (c) 2026 EterLine (Andrew)
// This file is part of fstmon.
// Licensed under the MIT License. See the LICENSE file for details.

package server

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"time"

	"github.com/eterline/ipcsv2base/internal/model"
)

/*
Server – HTTP server wrapper with optional TLS and graceful shutdown.

Holds underlying http.Server and provides Run and Close methods.
*/
type Server struct {
	srv     *http.Server
	tlsCfg  *tls.Config
	timeout time.Duration
	log     model.Logger
}

// ====================================================

// ServerOption – functional option type for configuring Server.
type ServerOption func(*Server)

// WithTLS – enables TLS with provided tls.Config.
func WithTLS(cfg *tls.Config) ServerOption {
	return func(s *Server) {
		s.tlsCfg = cfg
	}
}

// WithShutdownTimeout – sets shutdown timeout duration.
func WithShutdownTimeout(d time.Duration) ServerOption {
	return func(s *Server) {
		s.timeout = d
	}
}

// WithDisabledDefaultHttp2Map – disable default HTTP/2 map.
func WithDisabledDefaultHttp2Map() ServerOption {
	return func(s *Server) {
		s.srv.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler))
	}
}

// ====================================================

// NewServer – creates a new Server with given HTTP handler and options.
func NewServer(handler http.Handler, log model.Logger, opts ...ServerOption) *Server {
	s := &Server{
		srv: &http.Server{
			Handler: handler,
		},
		timeout: 5 * time.Second,
		log:     log,
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.tlsCfg != nil {
		s.srv.TLSConfig = s.tlsCfg
	}

	return s
}

/*
Run – starts the server on the given address.

	If TLS certificate and key paths are provided, runs TLS server.
	Listens for context cancellation to shutdown gracefully.
*/
func (s *Server) Run(ctx context.Context, addr, key, crt string) error {
	s.srv.Addr = addr
	tlsEnabled := s.srv.TLSConfig != nil && key != "" && crt != ""

	logger := s.log.With(
		model.FieldString("listen", addr),
		model.Field("tls", tlsEnabled),
	)

	if tlsEnabled {
		logger = logger.With(
			model.FieldString("ssl_cert", crt),
			model.FieldString("ssl_key", key),
		)
	}

	errCh := make(chan error, 1)

	go func() {
		var err error
		if tlsEnabled {
			err = s.srv.ListenAndServeTLS(crt, key)
		} else {
			err = s.srv.ListenAndServe()
		}
		errCh <- err
	}()

	logger.Info("server started")

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
		shCtx, cancel := context.WithTimeout(context.Background(), s.timeout)
		defer cancel()

		if err := s.srv.Shutdown(shCtx); err != nil {
			logger.Error("shutdown error", model.FieldError(err))
			return err
		}

		logger.Info("server stopped gracefully")
		return nil

	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			logger.Info("server closed normally")
			return nil
		}
		logger.Error("server error", model.FieldError(err))
		return err
	}
}

// Close – immediately shuts down the server with configured timeout.
func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	return s.srv.Shutdown(ctx)
}
