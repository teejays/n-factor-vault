package orm

import (
	"fmt"
)

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* M U T A T E
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// InsertOne inserts an entity into the DB
func InsertOne(v Entity) error {
	return gDB.Create(v).Error
}

// Save save an entity into the DB
func Save(v Entity) error {
	return gDB.Save(v).Error
}

// UpdateByColumn updates all the entities where the condition matches with the new Entity v
func UpdateByColumn(conditions map[string]interface{}, v Entity) error {
	db := gDB
	for col, val := range conditions {
		db = db.Where(fmt.Sprintf("%s = ?", col), val)
	}
	return db.Model(v).Updates(v).Error
}

// Delete soft deletes the entity v from  the DB
func Delete(v Entity) error {
	return gDB.Delete(v).Error
}
