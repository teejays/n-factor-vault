package server

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teejays/clog"
	api "github.com/teejays/n-factor-vault/backend/library/go-api"
)

func init() {
	clog.LogLevel = 0
}

func TestHandleSignup(t *testing.T) {

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
			wantErrMessage: "no content provided with the HTTP request",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Create the HTTP request and response
			var buff = bytes.NewBufferString(tt.content)
			var r = httptest.NewRequest(http.MethodPost, "/v1/signup", buff)
			var w = httptest.NewRecorder()

			// Call the Handler
			HandleSignup(w, r)

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
				assert.Contains(t, errH.Message, tt.wantErrMessage)

			}

		})
	}
}
