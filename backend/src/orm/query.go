package orm

import (
	"fmt"
	"sync"

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

func EmptyTable(table string) (int, error) {
	table = fmt.Sprintf("%s%s", gTableNamePrefix, table)
	result, err := gEngine.Exec(fmt.Sprintf("DELETE FROM %s WHERE 1=1", table))
	if err != nil {
		return -1, err
	}
	affected, err := result.RowsAffected()
	return int(affected), err
}

func EmptyTables(tables []string) error {
	for _, table := range tables {
		_, err := EmptyTable(table)
		if err != nil {
			return fmt.Errorf("could not empty %s: %v", table, err)
		}
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

	sess := gEngine.NewSession()

	gTestSessionLock.RLock()
	defer gTestSessionLock.RUnlock()
	if gTestSession != nil {
		sess = gTestSession
	}
	clog.Debugf("orm: Insert: value before insert:\n%+v", v)
	n, err := sess.InsertOne(v)
	if err != nil {
		return errWithContext(fmt.Errorf("could not save: %v\n%+v", err, v))
	}
	if n != 1 {
		// Case for panic?
		return errWithContext(fmt.Errorf("expected %d rows to be inserted but got %d", 1, n))
	}
	clog.Debugf("orm: Insert: value after insert:\n%+v", v)
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
