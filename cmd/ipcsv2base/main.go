// Copyright (c) 2026 EterLine (Andrew)
// This file is part of fstmon.
// Licensed under the MIT License. See the LICENSE file for details.

package main

import (
	"net/http"

	_ "net/http/pprof"

	"github.com/eterline/ipcsv2base/internal/app/ipcsv2base"
	"github.com/eterline/ipcsv2base/internal/config"
	"github.com/eterline/ipcsv2base/internal/infra/log"
	"github.com/eterline/ipcsv2base/internal/model"
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
		Profiling: "",
		Log: config.Log{
			LogLevel: "info",
			JSONlog:  false,
			Colored:  false,
		},
		Server: config.Server{
			Listen:     ":3000",
			CrtFileSSL: "",
			KeyFileSSL: "",
		},
		Base: config.Base{
			CountryTSV: []string{},
			CountryCSV: "",
			AsnCSV:     "",
			IPver:      "all",
		},
	}
)

func main() {
	var logapp model.Logger

	root := toolkit.InitAppStart(
		func() error {
			err := config.ParseArgs(&cfg)
			if err != nil {
				return err
			}

			logger, err := log.NewZapLoggerWithConfig(cfg.LogLevel, Flags.IsDev(), cfg.JSONlog, cfg.Colored)
			if err == nil {
				logapp = logger
			}
			return err
		},
	)

	if cfg.Profiling != "" {
		go func() {
			logapp.Info("pprof listening", model.FieldString("listen", cfg.Profiling))
			http.ListenAndServe(cfg.Profiling, nil)
		}()
	}

	ipcsv2base.Execute(root, logapp, Flags, cfg)
}
