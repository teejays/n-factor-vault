package orm

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/teejays/n-factor-vault/backend/library/id"
)

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* Q U E R Y
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

func FindByID(id id.ID, v Entity) (bool, error) {
	err := gDB.Where("id = ?", id).First(v).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func FindByColumn(colName string, colVal interface{}, v interface{}) (bool, error) {
	err := gDB.Where(map[string]interface{}{colName: colVal}).Find(v).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func FindOneByColumn(colName string, colVal interface{}, v Entity) (bool, error) {
	err := gDB.Where(map[string]interface{}{colName: colVal}).First(v).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func FindOne(conditions map[string]interface{}, v Entity) (bool, error) {
	db := gDB
	for col, val := range conditions {
		db = db.Where(fmt.Sprintf("%s = ?", col), val)
	}
	err := db.First(v).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func Find(conditions map[string]interface{}, v interface{}) (bool, error) {
	db := gDB
	for col, val := range conditions {
		db = db.Where(fmt.Sprintf("%s = ?", col), val)
	}
	err := db.Find(v).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
