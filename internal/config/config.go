// Copyright (c) 2026 EterLine (Andrew)
// This file is part of fstmon.
// Licensed under the MIT License. See the LICENSE file for details.

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alexflint/go-arg"
	"github.com/eterline/ipcsv2base/pkg/validate"
	"github.com/go-playground/validator/v10"
)

var val *validator.Validate

func init() {
	val = validator.New()
	val.RegisterStructValidation(baseStructValidation, Base{})
}

type (
	Log struct {
		LogLevel string `arg:"--log-level" help:"Logging level: debug|info|warn|error" validate:"oneof=debug info warn error"`
		JSONlog  bool   `arg:"--log-json,-j" help:"Set logs to JSON format"`
	}

	Server struct {
		Listen     string `arg:"--listen,-l" help:"Server listen address"`
		CrtFileSSL string `arg:"--ssl-cert,-c" help:"Server SSL certificate file" validate:"required_with=KeyFileSSL"`
		KeyFileSSL string `arg:"--ssl-key,-k" help:"Server SSL key file" validate:"required_with=CrtFileSSL"`
	}

	Base struct {
		CountryTSV []string `arg:"--country-tsvs" help:"Path to the country TSV files"`
		CountryCSV string   `arg:"--country-csv" help:"Path to the country CSV file"`
		AsnCSV     string   `arg:"--asn-csv" help:"Path to the ASN CSV file"`
		IPver      string   `arg:"--ip-ver" help:"IP version in base selector: all|v4|v6" validate:"oneof=all v4 v6"`
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

	if err := val.Struct(c); err != nil {
		switch errs := err.(type) {
		case validator.ValidationErrors:
			wrapped := validate.NewValidationErrorWrapper()
			for _, e := range errs {
				wrapped.Errors[e.StructNamespace()] = fmt.Sprintf("failed on '%s' tag", e.Tag())
			}
			return wrapped

		default:
			return fmt.Errorf("validation failed: %w", err)
		}
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
