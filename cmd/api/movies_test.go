package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
		body   string
		want   string
		code   int
	}{
		{
			name:   "create new movie",
			method: http.MethodPost,
			path:   "/v1/movies",
			body:   `{"title": "Moana", "runtime": 107, "genres": ["animation", "adventure"], "year": 2000}`,
			want:   "{Title:Moana Year:2000 Runtime:107 Genres:[animation adventure]}\n",
			code:   http.StatusCreated,
		},
		{
			name:   "create new movie badly-formed JSON 1",
			method: http.MethodPost,
			path:   "/v1/movies",
			body:   `?xml version="1.0" encoding="UTF-8"?><note><to>Alex</to></note>`,
			want:   `{"error":"body contains badly-formed JSON (at character 1)"}`,
			code:   http.StatusBadRequest,
		},
		{
			name:   "create new movie badly-formed JSON 2",
			method: http.MethodPost,
			path:   "/v1/movies",
			body:   `{"title": "Moana", }`,
			want:   `{"error":"body contains badly-formed JSON (at character 20)"}`,
			code:   http.StatusBadRequest,
		},
		{
			name:   "create new movie empty body",
			method: http.MethodPost,
			path:   "/v1/movies",
			want:   `{"error":"body must not be empty"}`,
			code:   http.StatusBadRequest,
		},
		{
			name:   "create new movie wrong JSON type",
			method: http.MethodPost,
			path:   "/v1/movies",
			body:   `{"title": 123}`,
			want:   `{"error":"body contains incorrect JSON type for field \"title\""}`,
			code:   http.StatusBadRequest,
		},
		{
			name:   "create new movie wrong method",
			method: http.MethodGet,
			path:   "/v1/movies",
			want:   `{"error":"the GET method is not supported for this resource"}`,
			code:   http.StatusMethodNotAllowed,
		},
		{
			name:   "get right movie id",
			method: http.MethodGet,
			path:   "/v1/movies/123",
			want:   `{"movie":{"id":123,"title":"Casablanca","runtime":"102 mins","genres":["drama","romance","war"],"version":1}}`,
			code:   http.StatusOK,
		},
		{
			name:   "get wrong movie id (negative)",
			method: http.MethodGet,
			path:   "/v1/movies/-112",
			want:   `{"error":"the requested resource could not be found"}`,
			code:   http.StatusNotFound,
		},
		{
			name:   "get wrong movie id (text)",
			method: http.MethodGet,
			path:   "/v1/movies/shrek",
			want:   `{"error":"the requested resource could not be found"}`,
			code:   http.StatusNotFound,
		},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%s", c.name), func(t *testing.T) {
			req, err := http.NewRequest(c.method, ts.URL+c.path, strings.NewReader(c.body))
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
