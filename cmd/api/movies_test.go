package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shortykevich/greenlight/internal/data"
)

func TestMovies(t *testing.T) {
	cfg := config{env: "development"}

	app := newApplication().
		setConfig(cfg).
		setLogger(nil).
		setModels(data.NewMockModels())

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
			name:   "create movie 1",
			method: http.MethodPost,
			path:   "/v1/movies",
			body:   `{"title":"Moana","year":2016,"runtime":"107 mins", "genres":["animation","adventure"]}`,
			want:   `{"movie":{"id":1,"title":"Moana","year":2016,"runtime":"107 mins","genres":["animation","adventure"],"version":1}}`,
			code:   http.StatusCreated,
		},
		{
			name:   "get movie",
			method: http.MethodGet,
			path:   "/v1/movies/1",
			want:   `{"movie":{"id":1,"title":"Moana","year":2016,"runtime":"107 mins","genres":["animation","adventure"],"version":1}}`,
			code:   http.StatusOK,
		},
		{
			name:   "create movie 2",
			method: http.MethodPost,
			path:   "/v1/movies",
			body:   `{"title":"Black Panther","year":2018,"runtime":"134 mins","genres":["action","adventure"]}`,
			want:   `{"movie":{"id":2,"title":"Black Panther","year":2018,"runtime":"134 mins","genres":["action","adventure"],"version":1}}`,
			code:   http.StatusCreated,
		},
		{
			name:   "create movie 3",
			method: http.MethodPost,
			path:   "/v1/movies",
			body:   `{"title":"Deadpool","year":2016, "runtime":"108 mins","genres":["action","comedy"]}`,
			want:   `{"movie":{"id":3,"title":"Deadpool","year":2016,"runtime":"108 mins","genres":["action","comedy"],"version":1}}`,
			code:   http.StatusCreated,
		},
		{
			name:   "create movie 4",
			method: http.MethodPost,
			path:   "/v1/movies",
			body:   `{"title":"The Breakfast Club","year":1986, "runtime":"96 mins","genres":["drama"]}`,
			want:   `{"movie":{"id":4,"title":"The Breakfast Club","year":1986,"runtime":"96 mins","genres":["drama"],"version":1}}`,
			code:   http.StatusCreated,
		},
		{
			name:   "delete movie",
			method: http.MethodDelete,
			path:   "/v1/movies/3",
			want:   `{"message:":"movie successfully deleted"}`,
			code:   http.StatusOK,
		},
		{
			name:   "get wrong movie",
			method: http.MethodGet,
			path:   "/v1/movies/3",
			want:   `{"error":"the requested resource could not be found"}`,
			code:   http.StatusNotFound,
		},
		{
			name:   "update movie",
			method: http.MethodPatch,
			path:   "/v1/movies/2",
			body:   `{"title":"Black Panther","year":2018,"runtime":"134 mins","genres":["sci-fi","action","adventure"]}`,
			want:   `{"movie":{"id":2,"title":"Black Panther","year":2018,"runtime":"134 mins","genres":["sci-fi","action","adventure"],"version":2}}`,
			code:   http.StatusCreated,
		},
		{
			name:   "update with empty body",
			method: http.MethodPatch,
			path:   "/v1/movies/2",
			want:   `{"error":"body must not be empty"}`,
			code:   http.StatusBadRequest,
		},
		{
			name:   "partial update",
			method: http.MethodPatch,
			path:   "/v1/movies/1",
			body:   `{"year":2000}`,
			want:   `{"movie":{"id":1,"title":"Moana","year":2000,"runtime":"107 mins","genres":["animation","adventure"],"version":2}}`,
			code:   http.StatusCreated,
		},
		{
			name:   "get all movies",
			method: http.MethodGet,
			path:   "/v1/movies",
			want:   `{"metadata":{},"movies":[{"id":1,"title":"Moana","year":2000,"runtime":"107 mins","genres":["animation","adventure"],"version":2},{"id":2,"title":"Black Panther","year":2018,"runtime":"134 mins","genres":["sci-fi","action","adventure"],"version":2},{"id":4,"title":"The Breakfast Club","year":1986,"runtime":"96 mins","genres":["drama"],"version":1}]}`,
			code:   http.StatusOK,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req, err := http.NewRequest(c.method, ts.URL+c.path, strings.NewReader(c.body))
			assertNoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			assertNoError(t, err)
			t.Cleanup(func() {
				_ = resp.Body.Close()
			})

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
