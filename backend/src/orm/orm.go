// Package orm follows a singleton pattern, which means that all
// packages will probably share the DB (for now).
package orm

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/go-xorm/xorm"
	"github.com/google/uuid"
	"xorm.io/core"

	// this is needed to support postgres connection
	_ "github.com/lib/pq"
	"github.com/teejays/clog"
	"github.com/teejays/n-factor-vault/backend/src/env"
)

var gEngine *xorm.Engine
var gDriverName = "postgres"

func init() {
	err := initEngine()
	if err != nil {
		clog.Fatalf("Could not get up new xorm.Engine: %v", err)
	}
}

func initEngine() error {

	connStr, err := getPostgresConnectionString()
	if err != nil {
		return err
	}
	clog.Debugf("xorm -> Postgres: connection string: %s", connStr)

	gEngine, err = xorm.NewEngine(gDriverName, connStr)
	if err != nil {
		return err
	}

	// Only set these settings if DEV
	if env.GetEnv() == env.DEV {
		gEngine.ShowSQL(true)
		gEngine.Logger().SetLevel(core.LOG_DEBUG)
	}

	return nil
}

func getPostgresConnectionString() (string, error) {

	// Get the port
	port, err := env.GetEnvVarInt("POSTGRES_PORT")
	if err != nil {
		return "", err
	}

	// Get the host
	host, err := env.GetEnvVar("POSTGRES_HOST")
	if err != nil {
		return "", err
	}

	// Get the database name
	dbName, err := env.GetEnvVar("POSTGRES_DBNAME")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("postgres://%s:%d/%s?sslmode=disable", host, port, dbName), nil
}

func getEnvVar(key string) (string, error) {
	val := os.Getenv(key)
	if strings.TrimSpace(val) == "" {
		return "", fmt.Errorf("env variable %s is not set or is empty", key)
	}
	return val, nil
}

func errWithContext(err error) error {
	if err != nil {
		err = fmt.Errorf("xorm error: %v", err)
	}
	return err
}

func RegisterModel(v interface{}) error {
	clog.Debugf("orm: syncing DB with type %v", reflect.TypeOf(v))
	// TODO: Do we need to ensure that v is of type pointer?
	err := gEngine.Sync2(v)
	if err != nil {
		return errWithContext(err)
	}
	return nil
}

var ErrNoRowsFound = fmt.Errorf("no rows found for the query")

func GetById(id uuid.UUID, v interface{}, must bool) error {
	has, err := gEngine.Table(v).Where("id = ?", id).Get(v)
	if err != nil {
		return errWithContext(err)
	}

	if must && !has {
		return errWithContext(ErrNoRowsFound)
	}

	return nil
}

func GetByColumn(columnName string, columnValue interface{}, v interface{}) (bool, error) {
	whereStmt := fmt.Sprintf("%s = ?", columnName)
	has, err := gEngine.Table(v).Where(whereStmt, columnValue).Get(v)
	if err != nil {
		return false, errWithContext(err)
	}

	return has, nil
}

func InsertOne(v interface{}) error {
	var err error
	clog.Debugf("Inserting:\n %+v\n", v)
	n, err := gEngine.InsertOne(v)
	if err != nil {
		return errWithContext(fmt.Errorf("could not save: %v\n%+v", err, v))
	}
	if n != 1 {
		// Case for panic.
		return errWithContext(fmt.Errorf("expected %d rows to be inserted but got %d", 1, n))
	}
	return nil
}
