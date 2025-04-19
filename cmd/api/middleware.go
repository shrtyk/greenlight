package main

import (
	"fmt"
	"net/http"

	"golang.org/x/time/rate"
)

/*
Wraps handler 'h' with 'mws' middlewares

IMPORTANT NOTE: The first middleware you list is the outermost wrapper, invoked BEFORE later ones
*/
func (app *application) applyMiddlewares(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for _, mw := range mws {
		h = mw(h)
	}
	return h
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	limiter := rate.NewLimiter(2, 4)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			app.rateLimitExceededResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
