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
	"github.com/teejays/n-factor-vault/backend/src/orm"
)

type AssertFunc func(t *testing.T, v interface{})

var AssertIsEqual = func(expected interface{}) AssertFunc {
	return func(t *testing.T, v interface{}) {
		assert.Equal(t, expected, v)
	}
}

var AssertNotEmptyFunc = func(t *testing.T, v interface{}) {
	assert.NotEmpty(t, v)
}

func EmptyTestTables(t *testing.T, tables []string) {
	if err := orm.EmptyTables(tables); err != nil {
		t.Fatalf("error emptying tables: %v", err)
	}
}

func init() {
	clog.LogLevel = 0
}

func TestHandleSignup(t *testing.T) {

	// There are two ways to make sure that while we run tests, the data in the database is not actually
	// persisted, otherwise it can affect the remaining tests.
	// 1. After each test, or run of a test, explicitly empty any relevant SQL table that the test might've written to
	// 2. Use orm.TestSession (usage shown below), which is like a transaction but makes sure that the transaction is
	// not committed. The problem with this approach is that some fields such as 'created_at', 'updated_at' etc. are
	// only populated in the Go instance of a struct once the struct is inserted/committed into the DB.
	// Exmample user of method (2):
	//
	// if err := orm.StartTestSession(); err != nil {
	// 	t.Errorf("could not start orm session: %v", err)
	// }
	// defer func() {
	// 	if err := orm.EndTestSession(); err != nil {
	// 		t.Errorf("could not end orm session: %v", err)
	// 	}
	// }()

	var relevantOrmTables = []string{"user_secure"}
	defer EmptyTestTables(t, relevantOrmTables)

	tests := []struct {
		name             string
		content          string
		wantStatusCode   int
		wantContent      string
		wantErrMessage   string
		assertFieldsJSON map[string]AssertFunc
		doNotEmptyTable  bool
	}{
		// TODO: Add test cases.
		{
			name:           "status BadRequest if request with empty content",
			content:        "",
			wantStatusCode: http.StatusBadRequest,
			wantContent:    "",
			wantErrMessage: "no content provided with the HTTP request",
		},
		{
			name:           "status BadRequest if request is not a valid JSON",
			content:        "I am a non-JSON content",
			wantStatusCode: http.StatusBadRequest,
			wantContent:    "",
			wantErrMessage: "content is not a valid JSON",
		},
		{
			name:           "status BadRequest if request does not include email and password",
			content:        `{}`,
			wantStatusCode: http.StatusBadRequest,
			wantContent:    "",
			wantErrMessage: "empty fields (name, email, password) provided",
		},
		{
			name:           "status BadRequest if request if name is missing",
			content:        `{"email":"email@email.com", "password":"secret"}`,
			wantStatusCode: http.StatusBadRequest,
			wantContent:    "",
			wantErrMessage: "empty fields (name) provided",
		},
		{
			name:           "status BadRequest if request if email is missing",
			content:        `{"name":"Jon Doe", "password":"secret"}`,
			wantStatusCode: http.StatusBadRequest,
			wantContent:    "",
			wantErrMessage: "empty fields (email) provided",
		},
		{
			name:           "status BadRequest if request if password is missing",
			content:        `{"name":"Jon Doe", "email":"email@email.com"}`,
			wantStatusCode: http.StatusBadRequest,
			wantContent:    "",
			wantErrMessage: "empty fields (password) provided",
		},
		{
			name:           "status BadRequest if email is not a valid email",
			content:        `{"name":"Jon Doe", "email":"email.email.com", "password":"secret"}`,
			wantStatusCode: http.StatusBadRequest,
			wantContent:    "",
			wantErrMessage: "email address has an invalid format",
		},
		{
			name:           "status OK if request has valid name, email, and password",
			content:        `{"name":"Jon Doe", "email":"email@email.com", "password":"secret"}`,
			wantStatusCode: http.StatusOK,
			wantContent:    "",
			wantErrMessage: "",
			assertFieldsJSON: map[string]AssertFunc{
				"id":         AssertNotEmptyFunc,
				"name":       AssertIsEqual("Jon Doe"),
				"email":      AssertIsEqual("email@email.com"),
				"created_at": AssertNotEmptyFunc,
				"updated_at": AssertNotEmptyFunc,
			},
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

			if tt.wantErrMessage != "" {
				var errH api.Error
				err = json.Unmarshal(body, &errH)
				if err != nil {
					t.Error(err)
				}
				assert.Equal(t, tt.wantStatusCode, int(errH.Code))
				assert.Contains(t, errH.Message, tt.wantErrMessage)

			}

			if tt.wantContent != "" {
				assert.Equal(t, tt.wantContent, string(body))
			}

			if tt.assertFieldsJSON != nil {
				var rJSON = make(map[string]interface{})
				err = json.Unmarshal(body, &rJSON)
				if err != nil {
					t.Error(err)
				}
				for k, assertFunc := range tt.assertFieldsJSON {
					v, exists := rJSON[k]
					if !exists {
						t.Errorf("the key '%s' does not exist in the response but an AssertFunc for it was specified", k)
					}
					assertFunc(t, v)
				}
			}

			// Empty the table unless specified
			if !tt.doNotEmptyTable {
				EmptyTestTables(t, relevantOrmTables)
			}

		})
	}
}
