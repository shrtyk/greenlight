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
	)

	server := app.routes()

	cases := []struct {
		name    string
		method  string
		path    string
		headers map[string][]string
		body    any
		want    envelope
		code    int
	}{
		{
			name:   "user creation",
			method: http.MethodPost,
			path:   "/v1/users",
			body: userCreateBody{
				Email:    "shortyk@example.com",
				Name:     "shortyk",
				Password: "pa55word",
			},
			want: envelope{
				"user": data.User{
					ID:        1,
					Email:     "shortyk@example.com",
					Name:      "shortyk",
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
				Email:    "shortyk@example.com",
				Name:     "shortyk",
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
				Email:    "shortyk@example.com",
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
				Email:    "shortyk@example.com",
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
				Email:    "shortyk@example.com",
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
				Email:    "shortyk",
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

	t.Run("test user authentication", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPost, "/v1/tokens/authentication", helpers.MustJSON(t, userAuthenticationBody{
			Email:    "shortyk@example.com",
			Password: "pa55word",
		}))
		assertions.AssertNoError(t, err)

		server.ServeHTTP(rw, req)

		plainText := app.models.Tokens.GetUserTokens(1).Authentication
		want, err := io.ReadAll(helpers.MustJSON(t, envelope{
			"authentication_token": data.Token{
				Plaintext: *plainText,
				Expiry:    data.MockTimeStamp,
			}}))
		assertions.AssertNoError(t, err)

		assertions.AssertStrings(t, rw.Body.String(), string(want))
	})

	t.Run("test user activation", func(t *testing.T) {
		rw := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodPut, "/v1/users/activated", helpers.MustJSON(t, activationToken{
			TokenPlainText: *app.models.Tokens.GetUserTokens(1).Activation,
		}))
		assertions.AssertNoError(t, err)

		server.ServeHTTP(rw, req)

		want, err := io.ReadAll(helpers.MustJSON(t, envelope{
			"user": data.User{
				ID:        1,
				Email:     "shortyk@example.com",
				Name:      "shortyk",
				CreatedAt: data.MockTimeStamp,
				Activated: true,
			},
		}))
		assertions.AssertNoError(t, err)

		assertions.AssertStrings(t, rw.Body.String(), string(want))
	})

}

func setRequestHeaders(t testing.TB, req *http.Request, headers map[string][]string) {
	t.Helper()

	for header, values := range headers {
		for _, val := range values {
			req.Header.Add(header, val)
		}
	}
}
