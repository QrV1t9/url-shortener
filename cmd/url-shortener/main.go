package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go-urlshortner/internal/config"
	"go-urlshortner/internal/http-server/handlers/url/deleter"
	"go-urlshortner/internal/http-server/handlers/url/redirect"
	"go-urlshortner/internal/http-server/handlers/url/save"
	mwLogger "go-urlshortner/internal/http-server/middleware/logger"
	"go-urlshortner/internal/lib/logger/handlers/slogpretty"
	"go-urlshortner/internal/lib/logger/sl"
	"go-urlshortner/internal/storage/postgres"
	"log/slog"
	"net/http"
	"os"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("Starting url-shortener", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	storage, err := postgres.New(cfg.StorageUrl)
	if err != nil {
		log.Error("failed to load storage", sl.Err(err))
		os.Exit(1)
	}
	defer storage.Close(context.Background())

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(mwLogger.New(log))

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shortener", map[string]string {
			cfg.HTTPServer.User: cfg.HTTPServer.Password,
		}))
		r.Post("/", save.New(log, storage))
		r.Delete("/{alias}", deleter.New(log, storage))
	})

	router.Get("/{alias}", redirect.New(log, storage))


	log.Info("starting server", slog.String("address", cfg.Address))

	srv := http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
