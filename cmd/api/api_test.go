package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shortykevich/greenlight/internal/data"
	"github.com/shortykevich/greenlight/internal/mailer"
	"github.com/shortykevich/greenlight/internal/testutils/assertions"
	"github.com/shortykevich/greenlight/internal/testutils/helpers"
	"github.com/shortykevich/greenlight/internal/validator"
)

func TestApi(t *testing.T) {
	cfg := config{env: "development"}
	models := data.NewMockModels()
	limiter := NewMockLimiter(false)

	mailData := &[]mailer.MailData{}
	mailer := mailer.NewMockMailer(mailData)

	app := newApplication(
		withConfig(cfg),
		withLogger(slog.Default()),
		withModels(models),
		withRateLimiter(limiter),
		withMailer(mailer),
		withVersion("test"),
	)

	server := app.routes()

	cases := []struct {
		name   string
		method string
		path   string
		body   any
		want   envelope
		code   int
	}{
		{
			name:   "user creation (bob)",
			method: http.MethodPost,
			path:   "/v1/users",
			body: userCreateBody{
				Email:    "bob@example.com",
				Name:     "bob",
				Password: "pa55word",
			},
			want: envelope{
				"user": data.User{
					ID:        1,
					Email:     "bob@example.com",
					Name:      "bob",
					CreatedAt: data.MockTimeStamp,
					Activated: false,
				},
			},
			code: http.StatusCreated,
		},
		{
			name:   "user creation (alice)",
			method: http.MethodPost,
			path:   "/v1/users",
			body: userCreateBody{
				Email:    "alice@example.com",
				Name:     "alice",
				Password: "pa55word",
			},
			want: envelope{
				"user": data.User{
					ID:        2,
					Email:     "alice@example.com",
					Name:      "alice",
					CreatedAt: data.MockTimeStamp,
					Activated: false,
				},
			},
			code: http.StatusCreated,
		},
		{
			name:   "user creation (tom)",
			method: http.MethodPost,
			path:   "/v1/users",
			body: userCreateBody{
				Email:    "tom@example.com",
				Name:     "tom",
				Password: "pa55word",
			},
			want: envelope{
				"user": data.User{
					ID:        3,
					Email:     "tom@example.com",
					Name:      "tom",
					CreatedAt: data.MockTimeStamp,
					Activated: false,
				},
			},
			code: http.StatusCreated,
		},
		{
			name:   "existed user creation",
			method: http.MethodPost,
			path:   "/v1/users",
			body: userCreateBody{
				Email:    "bob@example.com",
				Name:     "bob",
				Password: "pa55word",
			},
			want: envelope{
				"error": map[string]string{
					"email": "a user with this email address already exists",
				},
			},
			code: http.StatusUnprocessableEntity,
		},
		{
			name:   "user activation without provided token",
			method: http.MethodPut,
			path:   "/v1/users/activated",
			body: activationToken{
				TokenPlainText: "",
			},
			want: envelope{
				"error": map[string]string{
					"token": "must be provided",
				},
			},
			code: http.StatusUnprocessableEntity,
		},
		{
			name:   "user activation with long token",
			method: http.MethodPut,
			path:   "/v1/users/activated",
			body: activationToken{
				TokenPlainText: data.MockToken + "123",
			},
			want: envelope{
				"error": map[string]string{
					"token": "must be 26 bytes long",
				},
			},
			code: http.StatusUnprocessableEntity,
		},
		{
			name:   "user authentication with short password",
			method: http.MethodPost,
			path:   "/v1/tokens/authentication",
			body: userAuthenticationBody{
				Email:    "bob@example.com",
				Password: "a55word",
			},
			want: envelope{
				"error": map[string]string{
					"password": "must be at least 8 bytes long",
				},
			},
			code: http.StatusUnprocessableEntity,
		},
		{
			name:   "user authentication with wrong password",
			method: http.MethodPost,
			path:   "/v1/tokens/authentication",
			body: userAuthenticationBody{
				Email:    "bob@example.com",
				Password: "password",
			},
			want: envelope{
				"error": "invalid authentication credentials",
			},
			code: http.StatusUnauthorized,
		},
		{
			name:   "user authentication with empty password",
			method: http.MethodPost,
			path:   "/v1/tokens/authentication",
			body: userAuthenticationBody{
				Email:    "bob@example.com",
				Password: "",
			},
			want: envelope{
				"error": map[string]string{
					"password": "must be provided",
				},
			},
			code: http.StatusUnprocessableEntity,
		},
		{
			name:   "user authentication with wrong formated email",
			method: http.MethodPost,
			path:   "/v1/tokens/authentication",
			body: userAuthenticationBody{
				Email:    "bob",
				Password: "pa55word",
			},
			want: envelope{
				"error": map[string]string{
					"email": "must be a valid email address",
				},
			},
			code: http.StatusUnprocessableEntity,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rw := httptest.NewRecorder()
			req, err := http.NewRequest(c.method, c.path, helpers.MustJSON(t, c.body))
			assertions.AssertNoError(t, err)

			server.ServeHTTP(rw, req)

			want, err := io.ReadAll(helpers.MustJSON(t, c.want))
			assertions.AssertNoError(t, err)

			assertions.AssertStrings(t, rw.Body.String(), string(want))
			assertions.AssertStatusCode(t, rw.Code, c.code)
		})
	}

	// --------------------------------------------------------------------------------------------------------------

	authCases := []struct {
		name   string
		method string
		path   string
		userID int64
		body   any
		want   envelope
		code   int
	}{
		{
			name:   "bob authentication",
			method: http.MethodPost,
			path:   "/v1/tokens/authentication",
			userID: 1,
			body: userAuthenticationBody{
				Email:    "bob@example.com",
				Password: "pa55word",
			},
		},
		{
			name:   "alice authentication",
			method: http.MethodPost,
			path:   "/v1/tokens/authentication",
			userID: 2,
			body: userAuthenticationBody{
				Email:    "alice@example.com",
				Password: "pa55word",
			},
		},
		{
			name:   "tom authentication",
			method: http.MethodPost,
			path:   "/v1/tokens/authentication",
			userID: 3,
			body: userAuthenticationBody{
				Email:    "tom@example.com",
				Password: "pa55word",
			},
		},
	}

	for _, c := range authCases {
		t.Run(c.name, func(t *testing.T) {
			rw := httptest.NewRecorder()
			req, err := http.NewRequest(c.method, c.path, helpers.MustJSON(t, c.body))
			assertions.AssertNoError(t, err)

			server.ServeHTTP(rw, req)

			plainText := app.models.Tokens.GetUserTokens(c.userID).Authentication
			want, err := io.ReadAll(helpers.MustJSON(t, envelope{
				"authentication_token": data.Token{
					Plaintext: *plainText,
					Expiry:    data.MockTimeStamp,
				}}))
			assertions.AssertNoError(t, err)

			assertions.AssertStrings(t, rw.Body.String(), string(want))
		})
	}

	activationCases := []struct {
		name   string
		method string
		path   string
		userID int64
		want   envelope
		code   int
	}{
		{
			name:   "bob authentication",
			method: http.MethodPut,
			path:   "/v1/users/activated",
			userID: 1,
			want: envelope{
				"user": data.User{
					ID:        1,
					Email:     "bob@example.com",
					Name:      "bob",
					CreatedAt: data.MockTimeStamp,
					Activated: true,
				},
			},
		},
		{
			name:   "alice authentication",
			method: http.MethodPut,
			path:   "/v1/users/activated",
			userID: 2,
			want: envelope{
				"user": data.User{
					ID:        2,
					Email:     "alice@example.com",
					Name:      "alice",
					CreatedAt: data.MockTimeStamp,
					Activated: true,
				},
			},
		},
	}

	for _, c := range activationCases {
		t.Run(c.name, func(t *testing.T) {
			rw := httptest.NewRecorder()
			req, err := http.NewRequest(c.method, c.path, helpers.MustJSON(t, activationToken{
				TokenPlainText: *app.models.Tokens.GetUserTokens(c.userID).Activation,
			}))
			assertions.AssertNoError(t, err)

			server.ServeHTTP(rw, req)

			want, err := io.ReadAll(helpers.MustJSON(t, c.want))
			assertions.AssertNoError(t, err)

			assertions.AssertStrings(t, rw.Body.String(), string(want))
		})
	}

	// --------------------------------------------------------------------------------------------------------------

	bob, err := app.models.Users.GetByEmail("bob@example.com")
	assertions.AssertNoError(t, err)
	// alice, _ := app.models.Users.GetByEmail("alice@example.com")

	err = app.models.Permissions.AddForUser(bob.ID, data.MoviesWrite)
	assertions.AssertNoError(t, err)

	// At this point:
	// Bob has read and write permissions
	// Alice read only permissions
	// Tom account not activated

	bobAuthToken := *app.models.Tokens.GetUserTokens(1).Authentication
	aliceAuthToken := *app.models.Tokens.GetUserTokens(2).Authentication
	tomAuthToken := *app.models.Tokens.GetUserTokens(3).Authentication

	bobHeader := map[string][]string{
		"Authorization": {"Bearer " + bobAuthToken},
	}
	aliceHeader := map[string][]string{
		"Authorization": {"Bearer " + aliceAuthToken},
	}
	tomHeader := map[string][]string{
		"Authorization": {"Bearer " + tomAuthToken},
	}

	movieCases := []struct {
		name    string
		method  string
		path    string
		headers map[string][]string
		body    any
		want    envelope
		code    int
	}{
		{
			name:    "create movie 1",
			method:  http.MethodPost,
			path:    "/v1/movies",
			headers: bobHeader,
			body: movieCreateBody{
				Title:   "Moana",
				Year:    2016,
				Runtime: 107,
				Genres:  []string{"animation, adventure"},
			},
			want: envelope{
				"movie": data.Movie{
					ID:      1,
					Title:   "Moana",
					Year:    2016,
					Runtime: 107,
					Genres:  []string{"animation, adventure"},
					Version: 1,
				},
			},
			code: http.StatusCreated,
		},
		{
			name:    "get movie",
			method:  http.MethodGet,
			path:    "/v1/movies/1",
			headers: bobHeader,
			want: envelope{
				"movie": data.Movie{
					ID:      1,
					Title:   "Moana",
					Year:    2016,
					Runtime: 107,
					Genres:  []string{"animation, adventure"},
					Version: 1,
				},
			},
			code: http.StatusOK,
		},
		{
			name:    "get movie as non active user",
			method:  http.MethodGet,
			path:    "/v1/movies/1",
			headers: tomHeader,
			want: envelope{
				"error": "your user account must be activated to access this resource",
			},
			code: http.StatusForbidden,
		},
		{
			name:    "create movie 2",
			method:  http.MethodPost,
			path:    "/v1/movies",
			headers: bobHeader,
			body: movieCreateBody{
				Title:   "Black Panther",
				Year:    2018,
				Runtime: 134,
				Genres:  []string{"action", "adventure"},
			},
			want: envelope{
				"movie": data.Movie{
					ID:      2,
					Title:   "Black Panther",
					Year:    2018,
					Runtime: 134,
					Genres:  []string{"action", "adventure"},
					Version: 1,
				},
			},
			code: http.StatusCreated,
		},
		{
			name:    "create movie without write permission",
			method:  http.MethodPost,
			path:    "/v1/movies",
			headers: aliceHeader,
			body: movieCreateBody{
				Title:   "Black Panther",
				Year:    2018,
				Runtime: 134,
				Genres:  []string{"action", "adventure"},
			},
			want: envelope{
				"error": "your user account doesn't have the necessary permissions to access this resource",
			},
			code: http.StatusForbidden,
		},
		{
			name:    "create movie 3",
			method:  http.MethodPost,
			path:    "/v1/movies",
			headers: bobHeader,
			body: movieCreateBody{
				Title:   "Deadpool",
				Year:    2016,
				Runtime: 108,
				Genres:  []string{"action", "comedy"},
			},
			want: envelope{
				"movie": data.Movie{
					ID:      3,
					Title:   "Deadpool",
					Year:    2016,
					Runtime: 108,
					Genres:  []string{"action", "comedy"},
					Version: 1,
				},
			},
			code: http.StatusCreated,
		},
		{
			name:    "create movie 4",
			method:  http.MethodPost,
			path:    "/v1/movies",
			headers: bobHeader,
			body: movieCreateBody{
				Title:   "The Breakfast Club",
				Year:    1986,
				Runtime: 96,
				Genres:  []string{"drama"},
			},
			want: envelope{
				"movie": data.Movie{
					ID:      4,
					Title:   "The Breakfast Club",
					Year:    1986,
					Runtime: 96,
					Genres:  []string{"drama"},
					Version: 1,
				},
			},
			code: http.StatusCreated,
		},
		{
			name:    "delete movie 3 with permission",
			method:  http.MethodDelete,
			headers: bobHeader,
			path:    "/v1/movies/3",
			want:    envelope{"message:": "movie successfully deleted"},
			code:    http.StatusOK,
		},
		{
			name:    "delete non-existent movie",
			method:  http.MethodDelete,
			headers: bobHeader,
			path:    "/v1/movies/3",
			want:    envelope{"error": "the requested resource could not be found"},
			code:    http.StatusNotFound,
		},
		{
			name:    "update movie",
			method:  http.MethodPatch,
			path:    "/v1/movies/2",
			headers: bobHeader,
			body: getMovieUpdateBody(
				"Black Panther",
				2018,
				134,
				[]string{"sci-fi", "action", "adventure"},
			),
			want: envelope{
				"movie": data.Movie{
					ID:      2,
					Title:   "Black Panther",
					Year:    2018,
					Runtime: 134,
					Genres:  []string{"sci-fi", "action", "adventure"},
					Version: 2,
				},
			},
			code: http.StatusCreated,
		},
		{
			name:    "update with empty body",
			method:  http.MethodPatch,
			path:    "/v1/movies/2",
			body:    nil,
			headers: bobHeader,
			want:    envelope{"error": "body must not be empty"},
			code:    http.StatusBadRequest,
		},
		{
			name:    "partial update",
			method:  http.MethodPatch,
			path:    "/v1/movies/1",
			headers: bobHeader,
			body:    getMovieUpdateBody("", 2000, 0, []string{"animation"}),
			want: envelope{
				"movie": data.Movie{
					ID:      1,
					Title:   "Moana",
					Year:    2000,
					Runtime: 107,
					Genres:  []string{"animation"},
					Version: 2,
				},
			},
			code: http.StatusCreated,
		},
		{
			name:    "get all movies",
			method:  http.MethodGet,
			path:    "/v1/movies",
			headers: bobHeader,
			want: envelope{
				"metadata": data.Metadata{
					CurrentPage:  1,
					PageSize:     20,
					FirstPage:    1,
					LastPage:     1,
					TotalRecords: 3,
				},
				"movies": []data.Movie{
					{
						ID:      1,
						Title:   "Moana",
						Year:    2000,
						Runtime: 107,
						Genres:  []string{"animation"},
						Version: 2,
					},
					{
						ID:      2,
						Title:   "Black Panther",
						Year:    2018,
						Runtime: 134,
						Genres:  []string{"sci-fi", "action", "adventure"},
						Version: 2,
					},
					{
						ID:      4,
						Title:   "The Breakfast Club",
						Year:    1986,
						Runtime: 96,
						Genres:  []string{"drama"},
						Version: 1,
					},
				},
			},
			code: http.StatusOK,
		},
	}

	for _, c := range movieCases {
		t.Run(c.name, func(t *testing.T) {
			rw := httptest.NewRecorder()
			req, err := http.NewRequest(c.method, c.path, helpers.MustJSON(t, c.body))
			assertions.AssertNoError(t, err)

			setRequestHeaders(t, req, c.headers)

			server.ServeHTTP(rw, req)

			want, err := io.ReadAll(helpers.MustJSON(t, c.want))
			assertions.AssertNoError(t, err)

			assertions.AssertStrings(t, rw.Body.String(), string(want))
			assertions.AssertStatusCode(t, rw.Code, c.code)
		})
	}
}

func setRequestHeaders(t testing.TB, req *http.Request, headers map[string][]string) {
	t.Helper()

	for header, values := range headers {
		for _, val := range values {
			req.Header.Add(header, val)
		}
	}
}

func getMovieUpdateBody(title string, year int32, runtime data.Runtime, genres []string) movieUpdateBody {
	body := movieUpdateBody{
		Title:   nil,
		Year:    nil,
		Runtime: nil,
		Genres:  nil,
	}
	if len(title) != 0 {
		body.Title = &title
	}
	if year != 0 {
		body.Year = &year
	}
	if runtime != 0 {
		body.Runtime = &runtime
	}
	if genres != nil && validator.Unique(genres) {
		body.Genres = genres
	}
	return body
}
