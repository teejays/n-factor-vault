package server

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	api "github.com/teejays/n-factor-vault/backend/library/go-api"
)

var testEnvVariables = map[string]string{
	"ENV":           "testing",
	"POSTGRES_PORT": "5432",
	"POSTGRES_HOST": "localhost",
}

func setEnvVars(vars map[string]string) error {
	for k, v := range vars {
		if err := os.Setenv(k, v); err != nil {
			return err
		}
	}
	return nil
}

func unsetEnvVars(vars map[string]string) error {
	for k := range vars {
		if err := os.Unsetenv(k); err != nil {
			return err
		}
	}
	return nil
}

func unsetEnvVarsMust(t *testing.T, vars map[string]string) {
	if err := unsetEnvVars(testEnvVariables); err != nil {
		t.Fatalf("could not unset env variables at the end of test: %v", err)
	}
}

func TestHandleSignup(t *testing.T) {

	// Set the ENV vars for this test
	if err := setEnvVars(testEnvVariables); err != nil {
		t.Error(err)
	}
	defer unsetEnvVarsMust(t, testEnvVariables)

	tests := []struct {
		name           string
		content        string
		wantStatusCode int
		wantContent    string
		wantErrMessage string
	}{
		// TODO: Add test cases.
		{
			name:           "status bad request if empty content",
			content:        "",
			wantStatusCode: http.StatusBadRequest,
			wantContent:    "",
			wantErrMessage: "empty body",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Create the HTTP request and response
			var buff = bytes.NewBufferString(tt.content)
			var r = httptest.NewRequest(http.MethodPost, "/v1/signup", buff)
			var w = httptest.NewRecorder()

			// Call the Handler
			HandleLogin(w, r)

			// Verify the respoonse
			assert.Equal(t, tt.wantStatusCode, w.Code)

			resp := w.Result()
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Error(err)
			}
			if tt.wantErrMessage == "" {
				assert.Equal(t, tt.wantContent, string(body))
			} else {
				var errH api.Error
				err = json.Unmarshal(body, &errH)
				if err != nil {
					t.Error(err)
				}
				assert.Equal(t, tt.wantStatusCode, int(errH.Code))
				assert.True(t, strings.Contains(errH.Message, tt.wantErrMessage))

			}

		})
	}
}
