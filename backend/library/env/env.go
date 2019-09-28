package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/teejays/clog"
)

type AppEnv int

const (
	DEV AppEnv = iota
	STG
	PROD
	TEST
)

func (e AppEnv) String() string {

	switch e {
	case DEV:
		return "DEV"
	case STG:
		return "STG"
	case PROD:
		return "PROD"
	case TEST:
		return "TEST"
	default:
		return ""
	}

}

// GetAppEnv returns the DEV/STG/PROD environment that the code is running in.
func GetAppEnv() AppEnv {
	defaultEnv := DEV
	v, err := GetEnvVar("ENV")
	if err != nil {
		clog.Errorf("env: error getting env 'ENV': %v", err)
		return defaultEnv
	}
	v = strings.ToLower(v)
	if v == "prod" || v == "prd" || v == "production" {
		return PROD
	}
	if v == "stage" || v == "stg" || v == "staging" {
		return STG
	}
	if v == "dev" || v == "development" {
		return DEV
	}
	if v == "test" || v == "testing" {
		return TEST
	}
	return defaultEnv
}

// GetEnvVar returns the environment variables with key k. It errors if k is not setup or is empty.
func GetEnvVar(k string) (string, error) {
	val := os.Getenv(k)
	if strings.TrimSpace(val) == "" {
		return "", fmt.Errorf("env variable %s is not set or is empty", k)
	}
	return val, nil
}

// GetEnvVarInt returns the environment variables with key k as an int. It errors if k is not setup, is empty, or is not an int.
func GetEnvVarInt(k string) (int, error) {
	valStr, err := GetEnvVar(k)
	if err != nil {
		return 0, err
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return 0, fmt.Errorf("could not convert %s to int: %v", k, err)
	}
	return val, nil
}

// GetBoolOrDefault returns the true is an env variable is set to "true" or "1", false if set to "false" or "0"
// or default if is not set
func GetBoolOrDefault(k string, def bool) bool {
	val := os.Getenv(k)
	val = strings.TrimSpace(val)
	if val == "true" || val == "1" {
		return true
	}
	if val == "false" || val == "0" {
		return false
	}
	return def
}

func SetEnvVars(vars map[string]string) error {
	for k, v := range vars {
		clog.Debugf("orm: Setting env var %s to %s", k, v)
		if err := os.Setenv(k, v); err != nil {
			return err
		}
	}
	return nil
}

func SetEnvVarsMust(vars map[string]string) {
	if err := SetEnvVars(vars); err != nil {
		clog.Fatalf("could not set env vars: %v", err)
	}
}

func UnsetEnvVars(vars map[string]string) error {
	for k := range vars {
		clog.Debugf("orm: Unetting env var %s", k)
		if err := os.Unsetenv(k); err != nil {
			return err
		}
	}
	return nil
}

func UnsetEnvVarsMust(vars map[string]string) {
	if err := UnsetEnvVars(vars); err != nil {
		clog.Fatalf("could not unset env variables at the end of test: %v", err)
	}
}
