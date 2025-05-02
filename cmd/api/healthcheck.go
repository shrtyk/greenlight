package main

import (
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"enviroment": app.config.env,
			"version":    app.version,
		},
	}

	if err := app.writeJSON(w, env, http.StatusOK, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
