// Copyright (c) 2026 EterLine (Andrew)
// This file is part of fstmon.
// Licensed under the MIT License. See the LICENSE file for details.

package main

import (
	"github.com/eterline/ipcsv2base/internal/app/ipcsv2base"
	"github.com/eterline/ipcsv2base/internal/config"
	"github.com/eterline/ipcsv2base/internal/infra/log"
	"github.com/eterline/ipcsv2base/pkg/toolkit"
)

// -ladflags variables
var (
	CommitHash = "dev"
	Version    = "dev"
)

var (
	Flags = ipcsv2base.InitFlags{
		CommitHash: CommitHash,
		Version:    Version,
		Repository: "github.com/eterline/ipcsv2base",
	}

	cfg = config.Configuration{
		Log: config.Log{
			LogLevel:      "info",
			JSONlog:       false,
			AccessLogFile: "stdout",
		},
		Server: config.Server{
			Listen:     ":3000",
			CrtFileSSL: "",
			KeyFileSSL: "",
		},
	}
)

func main() {
	root := toolkit.InitAppStart(
		func() error {
			return config.ParseArgs(&cfg)
		},
	)

	logger := log.NewLogger(cfg.LogLevel, cfg.JSONlog)
	root.Context = log.WrapLoggerToContext(root.Context, logger)

	ipcsv2base.Execute(root, Flags, cfg)
}
