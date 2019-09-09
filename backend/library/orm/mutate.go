package orm

import (
	"fmt"
	"reflect"

	"github.com/teejays/n-factor-vault/backend/library/validator"
)

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* M U T A T E
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// InsertOne inserts an entity into the DB
func InsertOne(v Entity) error {
	var err error

	if v == nil {
		return fmt.Errorf("attempted to insert nil %s", reflect.TypeOf(v))
	}

	if reflect.ValueOf(v).IsNil() {
		return fmt.Errorf("attempted to insert nil %s", reflect.TypeOf(v))
	}

	// Run the validate struct based on validation tags
	err = validator.Validate(v)
	if err != nil {
		return err
	}

	// Run Entity Validators before we save a new entity
	validationErrs := v.ValidationErrors()
	if len(validationErrs) > 0 {
		return fmt.Errorf("validation failed: %v", validationErrs)
	}

	// Run BeforeCreate function for this entity
	err = v.BeforeCreate()
	if err != nil {
		return err
	}

	// Run BeforeSave function for this entity
	err = v.BeforeSave()
	if err != nil {
		return err
	}

	// Add to the DB
	err = gDB.Create(v).Error
	if err != nil {
		return err
	}

	// Run AfterCreate func
	err = v.AfterCreate()
	if err != nil {
		return err
	}

	// Run AfterSave func
	err = v.AfterSave()
	if err != nil {
		return err
	}

	return nil

}

// Save save an entity into the DB
func Save(v Entity) error {
	var err error

	// Run the validate struct based on validation tags
	err = validator.Validate(v)
	if err != nil {
		return err
	}

	// Run Validators before we save a new entity
	validationErrs := v.ValidationErrors()
	if len(validationErrs) > 0 {
		return fmt.Errorf("validation failed: %v", validationErrs)
	}

	// Run BeforeSave function for this entity
	err = v.BeforeSave()
	if err != nil {
		return err
	}

	// Save the entity in DB
	err = gDB.Save(v).Error
	if err != nil {
		return err
	}

	// Run AfterCreate func
	err = v.AfterCreate()
	if err != nil {
		return err
	}

	return nil
}

// UpdateByColumn updates all the entities where the condition matches with the new Entity v
// TODO: Refactor this function so we explicitly first fetch all the entities, mutate them,
// run all the Entity funcs, and then save.
func UpdateByColumn(conditions map[string]interface{}, v Entity) error {

	db := gDB
	for col, val := range conditions {
		db = db.Where(fmt.Sprintf("%s = ?", col), val)
	}
	return db.Model(v).Updates(v).Error
}

// Delete soft deletes the entity v from  the DB
func Delete(v Entity) error {
	var err error

	// Run BeforeDelete func
	err = v.BeforeDelete()
	if err != nil {
		return err
	}

	err = gDB.Delete(v).Error
	if err != nil {
		return err
	}

	// Run AfterDelete func
	err = v.AfterDelete()
	if err != nil {
		return err
	}

	return nil
}
