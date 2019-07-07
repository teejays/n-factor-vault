package apitest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teejays/n-factor-vault/backend/library/go-api"
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

type TestSuite struct {
	Route          string
	Method         string
	HandlerFunc    http.HandlerFunc
	AfterTestFunc  func(*testing.T)
	BeforeTestFunc func(*testing.T)
}

type HandlerTest struct {
	Name                string
	Content             string
	WantStatusCode      int
	WantContent         string
	WantErr             bool
	WantErrMessage      string
	AssertContentFields map[string]AssertFunc
	BeforeRunFunc       func(*testing.T)
	AfterRunFunc        func(*testing.T)
	SkipBeforeTestFunc  bool
	SkipAfterTestFunc   bool
}

type HandlerReqParams struct {
	Route               string
	Method              string
	Content             string
	HandlerFunc         http.HandlerFunc
	AcceptedStatusCodes []int
}

func MakeHandlerRequest(p HandlerReqParams) (*http.Response, []byte, error) {
	// Create the HTTP request and response
	var buff = bytes.NewBufferString(p.Content)
	var r = httptest.NewRequest(p.Method, p.Route, buff)
	var w = httptest.NewRecorder()

	// Call the Handler
	p.HandlerFunc(w, r)

	resp := w.Result()

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, body, err
	}

	// Check if the response status is one of the accepted ones
	if len(p.AcceptedStatusCodes) > 0 {
		var statusMap = make(map[int]bool)
		for _, status := range p.AcceptedStatusCodes {
			statusMap[status] = true
		}
		if v, hasKey := statusMap[w.Code]; !hasKey || !v {
			return resp, body, fmt.Errorf("apitest: handler request to %s resulted in a unaccepteable %d status:\n%s", p.Route, w.Code, string(body))
		}
	}

	return resp, body, nil
}

func (ts TestSuite) RunHandlerTests(t *testing.T, tests []HandlerTest) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			ts.RunHandlerTest(t, tt)
		})
	}
}
func (ts TestSuite) RunHandlerTest(t *testing.T, tt HandlerTest) {

	// Run BeforeRunFuncs
	if ts.BeforeTestFunc != nil && !tt.SkipBeforeTestFunc {
		ts.BeforeTestFunc(t)
	}

	if tt.BeforeRunFunc != nil {
		tt.BeforeRunFunc(t)
	}

	// Create the HTTP request and response
	hreq := HandlerReqParams{
		ts.Route,
		ts.Method,
		tt.Content,
		ts.HandlerFunc,
		[]int{tt.WantStatusCode},
	}
	resp, body, err := MakeHandlerRequest(hreq)
	assert.NoError(t, err)

	// Verify the respoonse
	assert.Equal(t, tt.WantStatusCode, resp.StatusCode)

	if tt.WantContent != "" {
		assert.Equal(t, tt.WantContent, string(body))
	}

	if tt.WantErrMessage != "" || tt.WantErr {
		var errH api.Error
		err = json.Unmarshal(body, &errH)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, tt.WantStatusCode, int(errH.Code))

		if tt.WantErr {
			assert.NotEmpty(t, errH.Message)
		}

		if tt.WantErrMessage != "" {
			assert.Contains(t, errH.Message, tt.WantErrMessage)
		}

	}

	// Run the individual assert functions for each of the field in the HTTP response body
	if tt.AssertContentFields != nil {
		// Unmarshall the body in to a map[string]interface{}
		var rJSON = make(map[string]interface{})
		err = json.Unmarshal(body, &rJSON)
		if err != nil {
			t.Error(err)
		}
		// Loop over all the available assert funcs specified and run them for the given field
		for k, assertFunc := range tt.AssertContentFields {
			v, exists := rJSON[k]
			if !exists {
				t.Errorf("the key '%s' does not exist in the response but an AssertFunc for it was specified", k)
			}
			assertFunc(t, v)
		}
	}

	// Run AfterRunFuncs
	if tt.AfterRunFunc != nil {
		tt.AfterRunFunc(t)
	}

	if ts.AfterTestFunc != nil && !tt.SkipAfterTestFunc {
		ts.AfterTestFunc(t)
	}

}
