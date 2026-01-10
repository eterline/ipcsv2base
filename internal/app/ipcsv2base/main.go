package ipcsv2base

import (
	"strings"
	"time"

	"github.com/eterline/ipcsv2base/internal/config"
	ipbaseProvide "github.com/eterline/ipcsv2base/internal/infra/ipbase"
	"github.com/eterline/ipcsv2base/internal/infra/log"
	"github.com/eterline/ipcsv2base/internal/interface/http/api"
	"github.com/eterline/ipcsv2base/internal/interface/http/baseapi"
	"github.com/eterline/ipcsv2base/internal/interface/http/server"
	"github.com/eterline/ipcsv2base/internal/service/ipbase"
	"github.com/eterline/ipcsv2base/pkg/toolkit"
	"github.com/go-chi/chi/v5"
)

func Execute(root *toolkit.AppStarter, flags InitFlags, cfg config.Configuration) {
	ctx := root.Context
	log := log.MustLoggerFromContext(ctx)

	// ========================================================

	log.Info(
		"starting app",
		"commit", flags.GetCommitHash(),
		"version", flags.GetVersion(),
		"repository", flags.GetRepository(),
	)

	defer func() {
		log.Info(
			"exit from app",
			"running_time", root.WorkTime(),
		)
	}()

	// ========================================================

	log.Info("starting IP base initialization")

	startInit := time.Now()

	// TODO: registry factory later...

	type Base interface {
		ipbase.MetaLookuper
		Size() int
	}

	var lookuper Base

	switch {
	case len(cfg.CountryTSV) > 0:
		log.Info(
			"loading CSV files",
			"tsv_files", strings.Join(cfg.CountryTSV, ", "),
		)

		base, err := ipbaseProvide.NewRegistryIPTSV(ctx, cfg.CountryTSV...)
		if err != nil {
			log.Error("failed to prepare IP base", "error", err)
			root.MustStopApp(1)
		}
		lookuper = base

	case cfg.CountryCSV != "" && cfg.AsnCSV != "":
		log.Info(
			"loading CSV files",
			"ip_version", cfg.Base.IPver,
			"asn_base", cfg.AsnCSV,
			"country_base", cfg.CountryCSV,
		)

		base, err := ipbaseProvide.NewRegistryIP(ctx, cfg.CountryCSV, cfg.AsnCSV, ipbaseProvide.IPVersionStr(cfg.IPver))
		if err != nil {
			log.Error("failed to prepare IP base", "error", err)
			root.MustStopApp(1)
		}
		lookuper = base

	default:
		log.Error("base did not prepared")
		root.MustStopApp(1)
	}

	log.Info(
		"ip base loaded successfully",
		"base_records", lookuper.Size(),
		"initialization_time_ms", time.Since(startInit).Milliseconds(),
	)

	baseSrvc := ipbase.NewIPBaseService(log, lookuper, &ipbaseProvide.IPbaseCacheMock{})
	baseHandlers := baseapi.NewBaseAPIHandlerGroup(log, baseSrvc, true)
	log.Info("base API handler group created")

	// ========================================================

	rootMux := chi.NewMux()
	rootMux.NotFound(api.HandleNotFound)
	rootMux.MethodNotAllowed(api.HandleNotAllowedMethod)

	// Types usage description
	rootMux.Get("/types", baseHandlers.AvailableTypes())

	rootMux.Route("/lookup", func(r chi.Router) {
		// Lookup by IP, path parameter or fallback to request IP
		r.Get("/ip/{ip}", baseHandlers.LookupIPHandler)
		r.Get("/ip/", baseHandlers.LookupIPHandler) // fallback: extract IP from request
		r.Get("/ip", baseHandlers.LookupIPHandler)  // fallback: extract IP from request
	})

	{
		srv := server.NewServer(
			rootMux,
			server.WithTLS(server.NewServerTlsConfig()),
			server.WithDisabledDefaultHttp2Map(),
		)

		root.WrapWorker(func() {
			err := srv.Run(ctx, cfg.Listen, cfg.KeyFileSSL, cfg.CrtFileSSL)
			if err != nil {
				log.Error("server run error", "error", err)
				root.StopApp()
			}
		})
		defer srv.Close()
	}

	// ============================
	root.WaitWorkers(15 * time.Second)
}
