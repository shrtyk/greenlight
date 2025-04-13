package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/shortykevich/greenlight/internal/data"
)

func (app *application) createMoviehandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "create a new movie")
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Casablanca",
		Runtime:   102,
		Genres:    []string{"drama", "romance", "war"},
		Version:   1,
	}

	if err := app.writeJson(w, envelope{"movie": movie}, http.StatusOK, nil); err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "The server encountered a problem and could not precoess your request", http.StatusInternalServerError)
	}
}
