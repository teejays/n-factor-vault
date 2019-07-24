package orm

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/teejays/clog"

	"github.com/teejays/n-factor-vault/backend/library/env"
	"github.com/teejays/n-factor-vault/backend/library/id"
)

// TODO: type DB {*gorm.DB} would just allows us to add more wrapper functions
// over *gorm.DB

var gDB *gorm.DB

func init() {
	connStr, err := getPostgresConnectionString()
	if err != nil {
		clog.Fatalf("Could not connect get postgres connection string: %v", err)
	}

	db, err := gorm.Open("postgres", connStr)
	if err != nil {
		clog.Fatalf("Could not connect to database: %v", err)
	}
	if env.GetEnv() == env.DEV || env.GetEnv() == env.TEST {
		db.LogMode(true)
	}
	gDB = db
	clog.Infof("orm: DB connection opened: %+v", gDB)
}

func getPostgresConnectionString() (string, error) {
	// Get the port
	port, err := env.GetEnvVarInt("POSTGRES_PORT")
	if err != nil {
		return "", err
	}

	// Get the host
	host, err := env.GetEnvVar("POSTGRES_HOST")
	if err != nil {
		return "", err
	}

	// Get the database name
	dbName, err := env.GetEnvVar("POSTGRES_DBNAME")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("host=%s port=%d dbname=%s sslmode=disable", host, port, dbName), nil
}

func RegisterModel(v interface{}) error {
	clog.Infof("orm: Registering model %T", v)
	return gDB.AutoMigrate(v).Error
}

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

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* Q U E R Y
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

func FindByID(id id.ID, v interface{}) (bool, error) {
	err := gDB.Where("id = ?", id).First(v).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func FindByColumn(colName string, colVal, v interface{}) (bool, error) {
	err := gDB.Where(map[string]interface{}{colName: colVal}).Find(v).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func FindOneByColumn(colName string, colVal, v interface{}) (bool, error) {
	err := gDB.Where(map[string]interface{}{colName: colVal}).First(v).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func FindOne(conditions map[string]interface{}, v interface{}) (bool, error) {
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
