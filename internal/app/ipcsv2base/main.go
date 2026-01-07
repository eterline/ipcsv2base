package ipcsv2base

import (
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
	log.Info("loading CSV files", "asn_base", cfg.AsnCSV, "country_base", cfg.CountryCSV)

	base, err := ipbaseProvide.NewRegistryIP(ctx, cfg.CountryCSV, cfg.AsnCSV)
	if err != nil {
		log.Error("Failed to prepare IP base", "error", err)
		root.MustStopApp(1)
	}

	log.Info("ip base loaded successfully", "base_records", base.Size())

	baseSrvc := ipbase.NewIPBaseService(log, base, &ipbaseProvide.IPbaseCacheMock{})
	baseHandlers := baseapi.NewBaseAPIHandlerGroup(log, baseSrvc, true)
	log.Info("base API handler group created")

	// ========================================================

	rootMux := chi.NewMux()
	rootMux.NotFound(api.HandleNotFound)
	rootMux.MethodNotAllowed(api.HandleNotAllowedMethod)

	rootMux.Route("/lookup", func(r chi.Router) {
		// Types usage description
		r.Get("/types", baseHandlers.AvailableTypes())
		// Lookup by IP, path parameter or fallback to request IP
		r.Get("/ip/{ip}", baseHandlers.LookupIPHandler)
		r.Get("/ip", baseHandlers.LookupIPHandler) // fallback: extract IP from request
		// Lookup by network prefix (subnet)
		r.Get("/subnet/{net}", baseHandlers.LookupSubnetHandler)
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
