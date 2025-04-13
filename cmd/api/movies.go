package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/shortykevich/greenlight/internal/data"
)

func (app *application) createMoviehandler(w http.ResponseWriter, r *http.Request) {
	movie := &struct {
		Title   string   `json:"title"`
		Year    int32    `json:"year"`
		Runtime int32    `json:"runtime"`
		Genres  []string `json:"genres"`
	}{}
	if err := app.readJson(w, r, movie); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "%+v\n", *movie)
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
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
		app.serverErrorResponse(w, r, err)
	}
}
