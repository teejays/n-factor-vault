package orm

import (
	"fmt"

	"github.com/teejays/clog"
)

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
	// by default, let's order by ID so the ordering is consistent across calls
	has, err := gEngine.Table(v).Where(whereStmt, columnValue).Asc("id").Get(v)
	if err != nil {
		return false, errWithContext(err)
	}

	return has, nil
}

func InsertOne(v interface{}) error {
	var err error
	clog.Debugf("Inserting:\n %+v\n", v)

	sess := gEngine.NewSession()

	// If we're in test mode and a test is using a test ORM session,
	// we should use that session instead
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
