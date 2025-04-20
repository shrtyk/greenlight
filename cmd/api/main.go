package main

import (
	"flag"
	"log/slog"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
	"github.com/shortykevich/greenlight/internal/data"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
}

func main() {
	var cfg config
	var rlCfg rateLimiterCfg

	flag.IntVar(&cfg.port, "port", 4000, "Api server port")
	flag.StringVar(&cfg.env, "env", "development", "Enviroment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "PostgreSQL max connection idle time")
	flag.Float64Var(&rlCfg.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&rlCfg.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&rlCfg.enable, "limiter-enabled", true, "Enable rate limiter")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	rateLimiter := NewRateLimiter(rlCfg)

	app := newApplication().
		setConfig(cfg).
		setLogger(logger).
		setRateLimiter(rateLimiter)

	db, err := app.openPostgresDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("couldn't close database", "err", err)
			os.Exit(1)
		}
	}()
	logger.Info("database connection pool established")

	app.setModels(data.NewModels(db))

	if err := app.server(); err != nil {
		app.logger.Error(err.Error())
		os.Exit(1)
	}
}
