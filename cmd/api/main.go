package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
	"github.com/shortykevich/greenlight/internal/data"
	"github.com/shortykevich/greenlight/internal/mailer"
)

func main() {
	var cfg config

	initFlags(&cfg)
	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()
	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		os.Exit(0)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	rateLimiter := NewRateLimiter(cfg.limiter)
	mailer := mailer.New(
		cfg.smtp.host,
		cfg.smtp.port,
		cfg.smtp.username,
		cfg.smtp.password,
		cfg.smtp.sender,
	)

	db, err := openPostgresDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer func() {
		if err = db.Close(); err != nil {
			logger.Error("couldn't close database", "err", err)
			os.Exit(1)
		}
	}()
	logger.Info("database connection pool established")

	app := newApplication(
		withConfig(cfg),
		withLogger(logger),
		withRateLimiter(rateLimiter),
		withMailer(mailer),
		withModels(data.NewModels(db)),
	)

	initBasicMetrics(db)

	if err = app.server(); err != nil {
		app.logger.Error(err.Error())
	}
}
