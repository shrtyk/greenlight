package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

type param struct {
	key string
	val string
}

func TestMovies(t *testing.T) {
	cfg := config{env: "development"}
	app := newApplication().
		setConfig(cfg).
		setLogger(nil)

	cases := []struct {
		name    string
		handler func(w http.ResponseWriter, r *http.Request)
		method  string
		params  *param
		url     string
		body    string
		code    int
		want    string
	}{
		{
			name:    "test create new movie",
			handler: app.createMoviehandler,
			method:  http.MethodPost,
			url:     "/v1/movies",
			body:    "",
			code:    http.StatusCreated,
			want:    "create a new movie\n",
		},
		{
			name:    "test get right movie id",
			handler: app.showMovieHandler,
			method:  http.MethodGet,
			params:  &param{"id", "123"},
			url:     "/v1/movies/123",
			code:    http.StatusOK,
			want:    "show the details of movie 123\n",
		},
		{
			name:    "test get wrong movie id",
			handler: app.showMovieHandler,
			method:  http.MethodGet,
			params:  &param{"id", "-112"},
			url:     "/v1/movies/-112",
			code:    http.StatusNotFound,
			want:    "404 page not found\n",
		},
		{
			name:    "test get wrong movie id (text)",
			handler: app.showMovieHandler,
			method:  http.MethodGet,
			params:  &param{"id", "shrek"},
			url:     "/v1/movies/shrek",
			code:    http.StatusNotFound,
			want:    "404 page not found\n",
		},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%s", c.name), func(t *testing.T) {
			req, err := http.NewRequest(c.method, c.url, nil)
			assertNoError(t, err)

			if c.params != nil {
				req = withRouterParams(req, c.params.key, c.params.val)
			}

			resp := httptest.NewRecorder()
			c.handler(resp, req)

			got, err := io.ReadAll(resp.Body)

			assertNoError(t, err)
			assertStatusCode(t, resp.Code, c.code)
			assertStrs(t, string(got), c.want)
		})
	}
}

func assertStatusCode(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("got status %v, want status %v", got, want)
	}
}

func withRouterParams(req *http.Request, key, value string) *http.Request {
	routerParams := httprouter.Params{
		httprouter.Param{Key: key, Value: value},
	}
	return req.WithContext(context.WithValue(req.Context(), httprouter.ParamsKey, routerParams))
}
