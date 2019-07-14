package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/teejays/clog"
)

// GetQueryParamInt extracts the param value with given name  out of the URL query
func GetQueryParamInt(r *http.Request, name string, defaultVal int) (int, error) {
	err := r.ParseForm()
	if err != nil {
		return defaultVal, err
	}
	values, exist := r.Form[name]
	clog.Debugf("URL values for %s: %+v", name, values)
	if !exist {
		return defaultVal, nil
	}
	if len(values) > 1 {
		return defaultVal, fmt.Errorf("multiple URL form values found for %s", name)
	}

	val, err := strconv.Atoi(values[0])
	if err != nil {
		return defaultVal, fmt.Errorf("error parsing %s value to an int: %v", name, err)
	}
	return val, nil
}

// GetMuxParamInt extracts the param with given name out of the route path
func GetMuxParamInt(r *http.Request, name string) (int64, error) {

	var vars = mux.Vars(r)
	clog.Debugf("MUX vars are: %+v", vars)
	valStr := vars[name]
	if strings.TrimSpace(valStr) == "" {
		return -1, fmt.Errorf("could not find var %s in the route", name)
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		return -1, fmt.Errorf("could not convert var %s to an int64: %v", name, err)
	}

	return int64(val), nil
}

// GetMuxParamStr extracts the param with given name out of the route path
func GetMuxParamStr(r *http.Request, name string) (string, error) {

	var vars = mux.Vars(r)
	clog.Debugf("MUX vars are: %+v", vars)
	valStr := vars[name]
	if strings.TrimSpace(valStr) == "" {
		return "", fmt.Errorf("var '%s' is not in the route", name)
	}

	return valStr, nil
}

// WriteResponse is a helper function to help write HTTP response
func WriteResponse(w http.ResponseWriter, code int, v interface{}) {
	writeResponse(w, code, v)
}

func writeResponse(w http.ResponseWriter, code int, v interface{}) {
	w.WriteHeader(code)
	clog.Debugf("api: writeResponse: content kind: %v; content:\n%+v", reflect.ValueOf(v).Kind(), v)

	if v == nil {
		return
	}

	// Json marshal the resp
	data, err := json.Marshal(v)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err, true, nil)
		return
	}

	// Write the response
	_, err = w.Write(data)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err, true, nil)
		return
	}
}

// WriteError is a helper function to help write HTTP response
func WriteError(w http.ResponseWriter, code int, err error, hide bool, overrideErr error) {
	writeError(w, code, err, hide, overrideErr)
}

func writeError(w http.ResponseWriter, code int, err error, hide bool, overrideErr error) {
	errMessage := CleanErrMessage(err.Error())
	clog.Error(errMessage)

	if hide {
		errMessage = ErrMessageClean
		if overrideErr != nil {
			errMessage = CleanErrMessage(overrideErr.Error())
		}
	}

	errE := NewError(code, errMessage)

	w.WriteHeader(code)
	data, err := json.Marshal(errE)
	if err != nil {
		panic(fmt.Sprintf("Failed to json.Unmarshal an error for http response: %v", err))
	}
	_, err = w.Write(data)
	if err != nil {
		panic(fmt.Sprintf("Failed to write error to the http response: %v", err))
	}
}

// UnmarshalJSONFromRequest takes in a pointer to an object and populates
// it by reading the content body of the HTTP request, and unmarshaling the
// body into the variable v.
func UnmarshalJSONFromRequest(r *http.Request, v interface{}) error {
	// Read the HTTP request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// api.WriteError(w, http.StatusBadRequest, err, false, nil)
		return err
	}
	defer r.Body.Close()

	if len(body) < 1 {
		// api.WriteError(w, http.StatusBadRequest, api.ErrEmptyBody, false, nil)
		return ErrEmptyBody
	}

	clog.Debugf("api: Unmarshalling to JSON: body:\n%+v", string(body))

	// Unmarshal JSON into Go type
	err = json.Unmarshal(body, &v)
	if err != nil {
		// api.WriteError(w, http.StatusBadRequest, err, true, api.ErrInvalidJSON)
		clog.Errorf("api: Error unmarshaling JSON: %v", err)
		return ErrInvalidJSON
	}

	return nil
}
