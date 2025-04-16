package main

import (
	"log/slog"

	"github.com/shortykevich/greenlight/internal/data"
)

type application struct {
	config config
	logger *slog.Logger
	models data.Models
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

func (app *application) setModels(models data.Models) *application {
	app.models = models
	return app
}
