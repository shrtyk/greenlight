package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
}

type application struct {
	config config
	logger *slog.Logger
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "Api server port")
	flag.StringVar(&cfg.env, "env", "development", "Enviroment (development|staging|production)")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	app := newApplication().
		setConfig(cfg).
		setLogger(logger)

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      mux,
		IdleTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	app.logger.Log(context.Background(), slog.LevelInfo, "starting server", "addr", server.Addr, "env", cfg.env)
	err := server.ListenAndServe()
	app.logger.Error(err.Error())
	os.Exit(1)
}

func newApplication() *application {
	return &application{}
}

func (app *application) setConfig(cfg config) *application {
	app.config = cfg
	return app
}

func (app *application) setLogger(log *slog.Logger) *application {
	app.logger = log
	return app
}
