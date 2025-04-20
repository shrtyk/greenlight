package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
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
	flag.Float64Var(&rlCfg.ratePerSecond, "rps-limit", 2, "Requests per second limit")
	flag.IntVar(&rlCfg.rateBurst, "req-burst-limit", 4, "Max amount of 'burst' requests")
	flag.BoolVar(&rlCfg.rateLimitEnabled, "limiter-on", true, "Limiter on/off")

	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	rateLimiter := NewRateLimiter(rlCfg)

	app := newApplication().
		setConfig(cfg).
		setLogger(logger).
		setRateLimiter(rateLimiter)

	go rateLimiter.innactiveClientsCleanUp()

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

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	app.logger.Log(context.Background(), slog.LevelInfo, "starting server", "addr", server.Addr, "env", cfg.env)
	err = server.ListenAndServe()
	app.logger.Error(err.Error())
	os.Exit(1)
}
