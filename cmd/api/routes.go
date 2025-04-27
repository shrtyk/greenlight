package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/v1/movies", app.requirePermission(app.listMoviesHandler, "movies:read"))
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.requirePermission(app.createMovieHandler, "movies:write"))
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.requirePermission(app.getMovieHandler, "movies:read"))
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requirePermission(app.updateMovieHandler, "movies:write"))
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requirePermission(app.deleteMovieHandler, "movies:write"))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	return app.applyMiddlewares(
		router,
		app.metrics,
		app.recoverPanic,
		app.enableCORS,
		app.rateLimit,
		app.authenticate,
	)
}
