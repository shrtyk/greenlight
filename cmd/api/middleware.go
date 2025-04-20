package main

import (
	"fmt"
	"net/http"
	"time"

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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.limiter.isEnabled() {
			next.ServeHTTP(w, r)
			return
		}

		ip, err := app.clientIP(r)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		app.limiter.mu.Lock()
		if _, ok := app.limiter.clients[ip]; !ok {
			app.limiter.clients[ip] = &client{
				limiter: rate.NewLimiter(app.limiter.getRPS(), app.limiter.getBurst()),
			}
		}

		app.limiter.clients[ip].lastSeen = time.Now()

		if !app.limiter.clients[ip].limiter.Allow() {
			app.limiter.mu.Unlock()
			app.rateLimitExceededResponse(w, r)
			return
		}
		app.limiter.mu.Unlock()

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
