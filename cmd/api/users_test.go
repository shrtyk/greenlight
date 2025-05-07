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

type userCreateBody struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type userUpdateBody struct {
	Email    string `json:"email,omitempty"`
	Name     string `json:"name,omitempty"`
	Password string `json:"password,omitempty"`
}

func TestUsers(t *testing.T) {
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
		name      string
		method    string
		path      string
		body      any
		want      envelope
		code      int
		lastToken string
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
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rw := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPost, "/v1/users", helpers.MustJSON(t, c.body))
			assertions.AssertNoError(t, err)

			server.ServeHTTP(rw, req)

			want, err := io.ReadAll(helpers.MustJSON(t, c.want))
			assertions.AssertNoError(t, err)

			assertions.AssertStrings(t, rw.Body.String(), string(want))
			assertions.AssertStatusCode(t, rw.Code, c.code)
		})
	}
}

func getLastActivationToken(data *[]mailer.MailData) string {
	ln := len(*data)
	if ln == 0 {
		return ""
	}
	lastToken := (*data)[ln-1]
	return lastToken.ActivationToken
}
