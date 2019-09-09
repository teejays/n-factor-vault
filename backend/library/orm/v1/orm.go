// Package orm_old follows a singleton pattern, which means that all
// packages will probably share the DB (for now).
package orm_old

import (
	"fmt"
	"reflect"

	"github.com/go-xorm/xorm"
	"github.com/teejays/clog"
	"xorm.io/core"

	// this is needed to support postgres connection
	_ "github.com/lib/pq"

	"github.com/teejays/n-factor-vault/backend/library/env"
)

var gEngine *xorm.Engine
var gDriverName = "postgres"
var gTableNamePrefix = "tb_"

func init() {
	err := initEngine(gDriverName)
	if err != nil {
		clog.Fatalf("Could not get up new xorm.Engine: %v", err)
	}
}

func initEngine(driverName string) error {
	clog.Debug("orm: Initializing ORM engine")
	connStr, err := getPostgresConnectionString()
	if err != nil {
		return err
	}
	clog.Debugf("xorm -> Postgres: connection string: %s", connStr)

	gEngine, err = xorm.NewEngine(driverName, connStr)
	if err != nil {
		return err
	}

	tbMapper := core.NewPrefixMapper(core.SnakeMapper{}, gTableNamePrefix)
	gEngine.SetTableMapper(tbMapper)
	gEngine.SetColumnMapper(core.GonicMapper{})

	// Only set these settings if DEV
	if env.GetAppEnv() == env.DEV {
		clog.Warnf("orm: setting high log level")
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

func errWithContext(err error) error {
	if err != nil {
		err = fmt.Errorf("xorm error: %v", err)
	}
	return err
}

func AutoMigrate(v interface{}) error {
	clog.Debugf("orm: syncing DB with type %v", reflect.TypeOf(v))
	// TODO: Do we need to ensure that v is of type pointer?
	err := gEngine.Cascade(true).Sync2(v)
	if err != nil {
		return errWithContext(err)
	}
	return nil
}
