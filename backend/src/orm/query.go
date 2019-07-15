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
	has, err := gEngine.Table(v).Where(whereStmt, columnValue).Get(v)
	if err != nil {
		return false, errWithContext(err)
	}

	return has, nil
}

func FindByColumn(columnName string, columnValue interface{}, result interface{}) error {
	whereStmt := fmt.Sprintf("%s = ?", columnName)
	// by default, let's order by ID so the ordering is consistent across calls
	err := gEngine.Where(whereStmt, columnValue).Asc("id").Find(result)
	if err != nil {
		return errWithContext(err)
	}

	return nil
}

func FindByColumns(whereColsVals map[string]interface{}, result interface{}) error {

	// Build the where statement
	var whereStmt string
	var vals []interface{}
	var cnt int
	for col, val := range whereColsVals {
		if cnt > 0 {
			whereStmt += " AND "
		}
		whereStmt += fmt.Sprintf("%s = ?", col)
		vals = append(vals, val)
		cnt++
	}
	clog.Debugf("orm: FindByCols: whereStmt: %s", whereStmt)

	// by default, let's order by ID so the ordering is consistent across calls
	err := gEngine.Where(whereStmt, vals...).Asc("id").Find(result)
	if err != nil {
		return errWithContext(err)
	}

	return nil
}

func InsertOne(v interface{}) error {
	var err error
	clog.Debugf("orm: InsertingOne:\n %+v\n", v)

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

func InsertMulti(vs interface{}) error {
	var err error
	clog.Debugf("orm: InsertingMulti:\n %+v\n", vs)

	sess := gEngine.NewSession()

	// If we're in test mode and a test is using a test ORM session,
	// we should use that session instead
	gTestSessionLock.RLock()
	defer gTestSessionLock.RUnlock()
	if gTestSession != nil {
		sess = gTestSession
	}

	clog.Debugf("orm: Insert: value before insert:\n%+v", vs)
	n, err := sess.InsertMulti(vs)
	if err != nil {
		return errWithContext(fmt.Errorf("could not save: %v\n%+v", err, vs))
	}
	clog.Debugf("orm: InsertMulti: %v rows inserted", n)
	return nil
}

// Update updates the rows that satisfies the conditions to value v
func Update(conditions map[string]string, v interface{}) error {
	var err error
	clog.Debugf("orm: Updating :\n %+v\n", v)

	sess := gEngine.NewSession()

	// If we're in test mode and a test is using a test ORM session,
	// we should use that session instead
	gTestSessionLock.RLock()
	defer gTestSessionLock.RUnlock()
	if gTestSession != nil {
		sess = gTestSession
	}

	clog.Debugf("orm: Update: value before update:\n%+v", v)

	for colName, colVal := range conditions {
		colStr := fmt.Sprintf("%s = ?", colName)
		sess = sess.Where(colStr, colVal)
	}
	n, err := sess.AllCols().Update(v)
	if err != nil {
		return errWithContext(fmt.Errorf("could not save: %v\n%+v", err, v))
	}
	clog.Debugf("orm: Update: %v rows affected", n)
	return nil
}
