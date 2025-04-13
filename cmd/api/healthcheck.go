package main

import (
	"net/http"
)

type healthReport struct {
	Status     string `json:"status"`
	Enviroment string `json:"enviroment"`
	Version    string `json:"version"`
}

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	data := healthReport{
		Status:     "available",
		Enviroment: app.config.env,
		Version:    version,
	}

	if err := app.writeJson(w, data, http.StatusOK, nil); err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}
}
