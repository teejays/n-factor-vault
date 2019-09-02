package orm

import (
	"fmt"
)

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* M U T A T E
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

func InsertOne(v interface{}) error {
	return gDB.Create(v).Error
}

func Save(v interface{}) error {
	return gDB.Save(v).Error
}

func UpdateByColumn(conditions map[string]interface{}, v interface{}) error {
	db := gDB
	for col, val := range conditions {
		db = db.Where(fmt.Sprintf("%s = ?", col), val)
	}
	return db.Model(v).Updates(v).Error
}

func Delete(v interface{}) error {
	return gDB.Delete(v).Error
}
