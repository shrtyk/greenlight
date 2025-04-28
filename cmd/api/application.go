package main

import (
	"context"
	"database/sql"
	"log/slog"
	"sync"
	"time"

	"github.com/shortykevich/greenlight/internal/data"
	"github.com/shortykevich/greenlight/internal/mailer"
)

type application struct {
	wg      sync.WaitGroup
	config  config
	logger  *slog.Logger
	models  data.Models
	limiter RateLimiter
	mailer  mailer.Mailer
}

type Option func(*application)

func newApplication(opts ...Option) *application {
	app := new(application)

	for _, opt := range opts {
		opt(app)
	}

	return app
}

func withConfig(cfg config) Option {
	return func(app *application) {
		app.config = cfg
	}
}

func withLogger(log *slog.Logger) Option {
	return func(app *application) {
		app.logger = log
	}
}

func withModels(models data.Models) Option {
	return func(app *application) {
		app.models = models
	}
}

func withRateLimiter(limiter RateLimiter) Option {
	return func(app *application) {
		app.limiter = limiter
	}
}

func withMailer(mailer mailer.Mailer) Option {
	return func(app *application) {
		app.mailer = mailer
	}
}

func openPostgresDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	db.SetConnMaxIdleTime(cfg.db.maxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}
