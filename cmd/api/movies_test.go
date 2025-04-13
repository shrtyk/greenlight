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

	router := app.routes()
	ts := httptest.NewServer(router)
	defer ts.Close()

	cases := []struct {
		name   string
		method string
		path   string
		want   string
		code   int
	}{
		{
			name:   "create new movie",
			method: http.MethodPost,
			path:   "/v1/movies",
			want:   "create a new movie\n",
			code:   http.StatusCreated,
		},
		{
			name:   "get right movie id",
			method: http.MethodGet,
			path:   "/v1/movies/123",
			want:   `{"movie":{"id":123,"title":"Casablanca","runtime":"102","genres":["drama","romance","war"],"version":1}}`,
			code:   http.StatusOK,
		},
		{
			name:   "get wrong movie id (negative)",
			method: http.MethodGet,
			path:   "/v1/movies/-112",
			want:   "404 page not found\n",
			code:   http.StatusNotFound,
		},
		{
			name:   "get wrong movie id (text)",
			method: http.MethodGet,
			path:   "/v1/movies/shrek",
			want:   "404 page not found\n",
			code:   http.StatusNotFound,
		},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%s", c.name), func(t *testing.T) {
			req, err := http.NewRequest(c.method, ts.URL+c.path, nil)
			assertNoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			assertNoError(t, err)
			defer resp.Body.Close()

			got, err := io.ReadAll(resp.Body)

			assertNoError(t, err)
			assertStatusCode(t, resp.StatusCode, c.code)
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
