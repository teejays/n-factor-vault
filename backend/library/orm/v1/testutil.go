package orm_old

import (
	"fmt"
	"sync"
	"testing"

	"github.com/go-xorm/xorm"
	"github.com/teejays/clog"
)

var gTestSession *xorm.Session
var gTestSessionLock sync.RWMutex

func StartTestSession() error {
	gTestSessionLock.Lock()
	defer gTestSessionLock.Unlock()
	if gTestSession != nil {
		return fmt.Errorf("orm: test session is already in use")
	}
	clog.Debugf("orm: Staring test session")
	gTestSession = gEngine.NewSession()

	return gTestSession.Begin()
}
func EndTestSession() error {
	gTestSessionLock.Lock()
	defer gTestSessionLock.Unlock()
	if gTestSession == nil {
		return fmt.Errorf("orm: test session is not in use, so can't end")
	}
	defer gTestSession.Close()
	err := gTestSession.Rollback()
	clog.Debugf("orm: Rolling back test session: %v", err)
	gTestSession = nil
	return err
}

func EmptyTable(objName string) (int, error) {
	table := gEngine.GetTableMapper().Obj2Table(objName)
	clog.Warnf("orm: emptying table %s", table)
	result, err := gEngine.Exec(fmt.Sprintf("DELETE FROM %s WHERE 1=1", table))
	if err != nil {
		return -1, err
	}
	affected, err := result.RowsAffected()
	return int(affected), err
}

func EmptyTables(objNames []string) error {
	for _, objName := range objNames {
		_, err := EmptyTable(objName)
		if err != nil {
			return fmt.Errorf("could not empty %s: %v", objName, err)
		}
	}
	return nil
}

func EmptyTestTables(t *testing.T, objNames []string) {
	if err := EmptyTables(objNames); err != nil {
		t.Fatalf("error emptying tables: %v", err)
	}
}
