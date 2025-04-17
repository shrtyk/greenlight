package main

import (
	"errors"
	"fmt"
	"net/http"

	data "github.com/shortykevich/greenlight/internal/data"
	"github.com/shortykevich/greenlight/internal/validator"
)

func (app *application) getMoviesHandler(w http.ResponseWriter, r *http.Request) {
	movies, err := app.models.Movies.GetAll()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if err := app.writeJson(w, envelope{"movies": movies}, http.StatusOK, nil); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	if err := app.readJson(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	v := validator.New()
	data.ValidateMovie(v, movie)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	if err := app.models.Movies.Insert(movie); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	if err := app.writeJson(w, envelope{"movie": movie}, http.StatusCreated, headers); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie, err := app.models.Movies.GetByID(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if err := app.writeJson(w, envelope{"movie": movie}, http.StatusOK, nil); err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie, err := app.models.Movies.GetByID(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	if err := app.readJson(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	movie.Title = input.Title
	movie.Year = input.Year
	movie.Runtime = input.Runtime
	movie.Genres = input.Genres

	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	if err := app.models.Movies.Update(movie); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJson(w, envelope{"movie": movie}, http.StatusCreated, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	if err := app.models.Movies.Delete(id); err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Following response is user friendly. Change response body to nil and status code to No Content if needed
	err = app.writeJson(w, envelope{"message:": "movie successfully deleted"}, http.StatusOK, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
