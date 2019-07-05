// Package orm follows a singleton pattern, which means that all
// packages will probably share the DB (for now).
package orm

import (
	"fmt"

	"github.com/go-xorm/xorm"
	"github.com/google/uuid"
	"xorm.io/core"

	// this is needed to support postgres connection
	_ "github.com/lib/pq"
	"github.com/teejays/clog"
)

var gEngine *xorm.Engine

var sampleDriverName = "postgres"
var sampleDataSource = "postgres://localhost:5432/nfactorvault?sslmode=disable"

func init() {
	var err error
	gEngine, err = xorm.NewEngine(sampleDriverName, sampleDataSource)
	if err != nil {
		clog.Fatalf("Could not get up new xorm.Engine: %v", err)
	}
	gEngine.ShowSQL(true)
	gEngine.Logger().SetLevel(core.LOG_DEBUG)
}

func errWithContext(err error) error {
	if err != nil {
		err = fmt.Errorf("xorm error: %v", err)
	}
	return err
}

func SyncModelSchema(v interface{}) error {

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
