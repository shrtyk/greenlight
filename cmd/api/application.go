package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/shortykevich/greenlight/internal/data"
	"github.com/shortykevich/greenlight/internal/mailer"
)

const version = "1.0.0"

type application struct {
	wg      sync.WaitGroup
	config  config
	logger  *slog.Logger
	models  data.Models
	limiter RateLimiter
	mailer  mailer.Mailer
}

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
	limiter rateLimiterCfg
	smtp    struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	cors struct {
		trustedOrigins []string
	}
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

func initFlags(cfg *config) {
	flag.IntVar(&cfg.port, "port", 4000, "Api server port")
	flag.StringVar(&cfg.env, "env", "development", "Enviroment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enable, "limiter-enabled", true, "Enable rate limiter")
	flag.DurationVar(
		&cfg.limiter.cleanupFreq,
		"clean-clients-freq",
		3*time.Minute,
		"Frequency of cleaning up limiter cache",
	)
	flag.DurationVar(
		&cfg.limiter.rebuildFreq,
		"rebuild-clients-freq",
		6*time.Hour,
		"Frequency of rebuilding limiter cache to prevent map memory leak",
	)

	flag.StringVar(&cfg.smtp.host, "smtp-host", os.Getenv("SMTP_HOST"), "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", os.Getenv("SMTP_USERNAME"), "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", os.Getenv("SMTP_PASSWORD"), "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", os.Getenv("SMTP_SENDER"), "SMTP sender")

	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(origin string) error {
		cfg.cors.trustedOrigins = strings.Fields(origin)
		return nil
	})
}

func initBasicMetrics(database *sql.DB) {
	expvar.NewString("version").Set(version)
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))
	expvar.Publish("database", expvar.Func(func() any {
		return database.Stats()
	}))
	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))
}
