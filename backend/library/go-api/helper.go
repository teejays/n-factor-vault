package api

import (
	"encoding/json"
	"fmt"
	"net/http"
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

// GetMuxParamrInt extracts the param with given name out of the route path
func GetMuxParamrInt(r *http.Request, name string) (int64, error) {

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

// WriteResponse is a helper functoin to help write HTTP response
func WriteResponse(w http.ResponseWriter, code int, v interface{}) {
	writeResponse(w, code, v)
}

func writeResponse(w http.ResponseWriter, code int, v interface{}) {
	w.WriteHeader(code)

	if v == nil {
		return
	}

	// Json marshal the resp
	data, err := json.Marshal(v)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err, true)
		return
	}
	// Write the response
	_, err = w.Write(data)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err, true)
		return
	}
}

// WriteError is a helper functoin to help write HTTP response
func WriteError(w http.ResponseWriter, code int, err error, hide bool) {
	writeError(w, code, err, hide)
}

func writeError(w http.ResponseWriter, code int, err error, hide bool) {
	errMessage := CleanErrMessage(err.Error())
	clog.Error(errMessage)

	if hide {
		errMessage = ErrMessageClean
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
