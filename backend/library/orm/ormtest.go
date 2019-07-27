package orm

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/teejays/clog"
)

func emptyTable(model interface{}) (int, error) {
	clog.Warnf("orm: emptying table for %s", reflect.ValueOf(model).Type())
	db := gDB.Unscoped().Delete(model)
	return int(db.RowsAffected), db.Error
}

func EmptyTables(models ...interface{}) error {
	// clog.Debugf("orm: EmptyTables(): param type: %d %s", reflect.TypeOf(models), reflect.ValueOf(models).Kind())
	for _, m := range models {
		// clog.Debugf("orm: EmptyTables(): Processing model #%d: %s", i+1, reflect.TypeOf(m))
		// clog.Debugf("orm: EmptyTestTables(): Processing model #%d: %s", i+1, reflect.ValueOf(m).Elem().Type())
		n, err := emptyTable(m)
		clog.Debugf("orm: EmptyTables: rows deleted %d", n)
		if err != nil {
			return fmt.Errorf("could not empty %s: %v", getTableName(m), err)
		}
	}
	return nil
}

func EmptyTestTables(t *testing.T, models ...interface{}) {

	// clog.Debugf("orm: EmptyTestTables(): Number of models: %d", len(models))
	// clog.Debugf("orm: EmptyTestTables(): param type: %s %s", reflect.TypeOf(models), reflect.ValueOf(models).Kind())
	if err := EmptyTables(models...); err != nil {
		t.Fatalf("error emptying tables: %v", err)
	}
}

func getTableName(model interface{}) string {
	clog.Debugf("orm: getTableName(): model: %v", reflect.ValueOf(model).Type())
	// clog.Debugf("orm: getTableName(): gDB: %+v", gDB)
	scope := gDB.NewScope(model)
	scope.Search = nil
	//clog.Debugf("orm: getTableName(): scope: %+v", scope)
	tableName := scope.TableName()
	//clog.Debugf("orm: getTableName(): tableName: %s", tableName)
	return tableName
}
