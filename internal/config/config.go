// Copyright (c) 2026 EterLine (Andrew)
// This file is part of fstmon.
// Licensed under the MIT License. See the LICENSE file for details.

package config

import (
	"os"
	"path/filepath"

	"github.com/alexflint/go-arg"
)

type (
	Log struct {
		LogLevel      string `arg:"--log-level" help:"Logging level: debug|info|warn|error"`
		JSONlog       bool   `arg:"--log-json,-j" help:"Set logs to JSON format"`
		AccessLogFile string `arg:"--access-log" help:"Set access log file"`
	}

	Server struct {
		Listen     string `arg:"--listen,-l" help:"Server listen address"`
		CrtFileSSL string `arg:"--ssl-cert,-c" help:"Server SSL certificate file"`
		KeyFileSSL string `arg:"--ssl-key,-k" help:"Server SSL key file"`
	}

	Base struct {
		CountryCSV string `arg:"--country-csv" help:"Path to the country CSV file"`
		AsnCSV     string `arg:"--asn-csv" help:"Path to the ASN CSV file"`
	}

	Configuration struct {
		Profiling string `arg:"--prof-listen" help:"pprof server listen address"`
		Log
		Server
		Base
	}
)

var (
	parserConfig = arg.Config{
		Program:           selfExec(),
		IgnoreEnv:         false,
		IgnoreDefault:     false,
		StrictSubcommands: true,
	}
)

func ParseArgs(c *Configuration) error {
	p, err := arg.NewParser(parserConfig, c)
	if err != nil {
		return err
	}

	err = p.Parse(os.Args[1:])
	if err == arg.ErrHelp {
		p.WriteHelp(os.Stdout)
		os.Exit(1)
	}
	return err
}

func selfExec() string {
	exePath, err := os.Executable()
	if err != nil {
		return "monita"
	}

	return filepath.Base(exePath)
}
