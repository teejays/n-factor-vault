package apitest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teejays/n-factor-vault/backend/library/go-api"
)

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* T E S T   S U I T E
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// TestSuite defines a configuration that wraps a bunch of individual tests for a single HandlerFunc
type TestSuite struct {
	Route                 string
	Method                string
	HandlerFunc           http.HandlerFunc
	Handler               http.Handler
	AuthBearerTokenFunc   func(*testing.T) string
	AuthMiddlewareHandler api.MiddlewareFunc
	AfterTestFunc         func(*testing.T)
	BeforeTestFunc        func(*testing.T)
}

// HandlerTest defines configuration for a single test run for a HandlerFunc. It is run run as part of the TestSuite
type HandlerTest struct {
	Name    string
	Content string

	BeforeRunFunc       func(*testing.T)
	AfterRunFunc        func(*testing.T)
	AuthBearerTokenFunc func(*testing.T) string
	SkipAuthToken       bool
	SkipBeforeTestFunc  bool
	SkipAfterTestFunc   bool

	WantStatusCode      int
	WantContent         string
	AssertContentFields map[string]AssertFunc // This only works if the response if a map (not if it's an array)
	AssertContentFuncs  []AssertFunc
	WantErr             bool
	WantErrMessage      string
}

// RunHandlerTests runs all the HandlerTests inside a testing.T.Run() loop
func (ts TestSuite) RunHandlerTests(t *testing.T, tests []HandlerTest) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			ts.RunHandlerTest(t, tt)
		})
	}
}

// RunHandlerTest run all the HandlerTest tt
func (ts TestSuite) RunHandlerTest(t *testing.T, tt HandlerTest) {
	// If both HandlerFunc and Handler are provided, we don't know which one to use
	if ts.HandlerFunc != nil && ts.Handler != nil {
		t.Errorf("both Handler and HandlerFunc field provided in TestSuite for %s", ts.Route)
		return
	}

	// Run BeforeRunFuncs
	if ts.BeforeTestFunc != nil && !tt.SkipBeforeTestFunc {
		ts.BeforeTestFunc(t)
	}

	if tt.BeforeRunFunc != nil {
		tt.BeforeRunFunc(t)
	}

	// Figure out what handler are we using
	handler := ts.Handler
	if handler == nil {
		handler = ts.HandlerFunc
	}
	if handler == nil {
		t.Errorf("both Handler and HandlerFunc provided in TestSuite are nil for %s", ts.Route)
	}

	// Authorization?
	if ts.AuthMiddlewareHandler != nil {
		handler = ts.AuthMiddlewareHandler(handler)
	}

	var authBearerToken string
	// If we have specified an AuthBearerTokenFunc, use it. If there is one in tt as well, use that one.
	if ts.AuthBearerTokenFunc != nil && !tt.SkipAuthToken {
		authBearerToken = ts.AuthBearerTokenFunc(t)
	}
	if tt.AuthBearerTokenFunc != nil && !tt.SkipAuthToken {
		authBearerToken = tt.AuthBearerTokenFunc(t)
	}

	// Create the HTTP request and response
	hreq := HandlerReqParams{
		Route:           ts.Route,
		Method:          ts.Method,
		Handler:         handler,
		AuthBearerToken: authBearerToken,
	}
	resp, body, err := hreq.MakeHandlerRequest(tt.Content, nil)
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

	if tt.AssertContentFuncs != nil {
		// Unmarshall the body in to a map[string]interface{}
		var rJSON interface{}
		err = json.Unmarshal(body, &rJSON)
		if err != nil {
			t.Error(err)
		}
		for _, f := range tt.AssertContentFuncs {
			f(t, rJSON)
		}
	}

	if t.Failed() {
		fmt.Printf("apitest: Content validation failed for:\n%s\n", body)
	}

	// Run AfterRunFuncs
	if tt.AfterRunFunc != nil {
		tt.AfterRunFunc(t)
	}

	if ts.AfterTestFunc != nil && !tt.SkipAfterTestFunc {
		ts.AfterTestFunc(t)
	}

}

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* H A N D L E R   R E Q U E S T
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// HandlerReqParams define a set of configuration that allow us to make repeated calls to Handler
type HandlerReqParams struct {
	Route  string
	Method string
	// Only one of HandlerFunc or Handler should be provided
	HandlerFunc     http.HandlerFunc
	Handler         http.Handler
	AuthBearerToken string
	Middlewares     []api.MiddlewareFunc
}

// MakeHandlerRequest makes an request to the handler specified in p, using the content. It errors if there is an
// error making the request, or if the received status code is not among the accepted status codes
func (p HandlerReqParams) MakeHandlerRequest(content string, acceptedStatusCodes []int) (*http.Response, []byte, error) {

	// Figure out what handler are we using
	handler := p.Handler
	if handler == nil {
		handler = p.HandlerFunc
	}
	// If both HandlerFunc and Handler are provided, we don't know which one to use
	if handler == nil {
		return nil, nil, fmt.Errorf("both Handler and HandlerFunc field provided in HandlerReqParams for %s", p.Route)
	}

	// Create the HTTP request and response
	var buff = bytes.NewBufferString(content)
	var r = httptest.NewRequest(p.Method, p.Route, buff)
	var w = httptest.NewRecorder()

	// Add Authenticate header to request
	if p.AuthBearerToken != "" {
		r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", p.AuthBearerToken))
	}

	// Add Middlewares
	for _, mw := range p.Middlewares {
		handler = mw(handler)
	}

	// Call the Handler
	handler.ServeHTTP(w, r)

	resp := w.Result()

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, body, err
	}

	// Check if the response status is one of the accepted ones
	if len(acceptedStatusCodes) > 0 {
		var statusMap = make(map[int]bool)
		for _, status := range acceptedStatusCodes {
			statusMap[status] = true
		}
		if v, hasKey := statusMap[w.Code]; !hasKey || !v {
			return resp, body, fmt.Errorf("apitest: handler request to %s resulted in a unaccepteable %d status:\n%s", p.Route, w.Code, string(body))
		}
	}

	return resp, body, nil
}

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* A S S E R T   F U N C S
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// AssertFunc is a function that takes the testing.T pointer, a value v, and asserts
// whether v is good
type AssertFunc func(t *testing.T, v interface{})

// AssertIsEqual is a of type AssertFunc. It verifies that the value v is equal to the expected value.
var AssertIsEqual = func(expected interface{}) AssertFunc {
	return func(t *testing.T, v interface{}) {
		assert.Equal(t, expected, v)
	}
}

// AssertNotEmptyFunc is a of type AssertFunc. It verifies that the value v is not empty.
var AssertNotEmptyFunc = func(t *testing.T, v interface{}) {
	assert.NotEmpty(t, v)
}

// AssertIsSlice asserts that v is a slice or an array
var AssertIsSlice = func(t *testing.T, v interface{}) {
	if _, ok := v.([]interface{}); !ok {
		t.Errorf("could not assert that it's a slice, it's %s", reflect.ValueOf(v).Kind())
	}
}

// AssertSliceOfLen asserts that v is a slice or an array with n elements
var AssertSliceOfLen = func(n int) AssertFunc {
	return func(t *testing.T, v interface{}) {
		_v, ok := v.([]interface{})
		if !ok {
			t.Errorf("could not assert that it's a slice, it's %s", reflect.ValueOf(v).Kind())
			return
		}
		assert.Equal(t, n, len(_v))
	}
}
