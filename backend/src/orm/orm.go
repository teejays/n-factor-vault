// Package orm follows a singleton pattern, which means that all
// packages will probably share the DB (for now).
package orm

import (
	"fmt"
	"reflect"

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

	tbMapper := core.NewPrefixMapper(core.SnakeMapper{}, "tb_")
	gEngine.SetTableMapper(tbMapper)
	gEngine.SetColumnMapper(core.GonicMapper{})

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

func errWithContext(err error) error {
	if err != nil {
		err = fmt.Errorf("xorm error: %v", err)
	}
	return err
}

func RegisterModel(v interface{}) error {
	clog.Debugf("orm: syncing DB with type %v", reflect.TypeOf(v))
	// TODO: Do we need to ensure that v is of type pointer?
	err := gEngine.Cascade(true).Sync2(v)
	if err != nil {
		return errWithContext(err)
	}
	return nil
}

var ErrNoRowsFound = fmt.Errorf("no rows found for the query")

func GetByID(id ID, v interface{}) (bool, error) {
	has, err := gEngine.Table(v).Where("id = ?", id).Get(v)
	if err != nil {
		return false, errWithContext(err)
	}

	return has, nil
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

func InsertTx(vs ...interface{}) (err error) {

	clog.Debugf("Inserting:\n %+v\n", vs...)

	sess := gEngine.NewSession()
	defer sess.Close()

	err = sess.Begin()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			sess.Rollback()
			return
		}
		err = sess.Commit()
		if err != nil {
			clog.Errorf("orm: error while committing insert transaction: %v\nRolling back transaction...", err)
			sess.Rollback()
			return
		}
	}()

	n, err := sess.Insert(vs...)
	if err != nil {
		err = errWithContext(fmt.Errorf("could not save: %v\n%+v", err, vs))
		return
	}
	if n != int64(len(vs)) {
		// Case for panic?
		err = errWithContext(fmt.Errorf("expected %d rows to be inserted but got %d", 1, n))
		return
	}
	return
}

func GetNewID() ID {
	return ID(uuid.New().String())
}
