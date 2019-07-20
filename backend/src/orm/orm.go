package orm

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/teejays/clog"
	"github.com/teejays/n-factor-vault/backend/library/env"
	"github.com/teejays/n-factor-vault/backend/library/id"
)

var gDBConn *gorm.DB

func init() {
	connStr, err := getPostgresConnectionString()
	if err != nil {
		clog.Fatalf("Could not connect get postgres connection string: %v", err)
	}

	db, err := gorm.Open("postgres", connStr)
	if err != nil {
		clog.Fatalf("Could not connect to database: %v", err)
	}
	gDBConn = db
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

func AutoMigrate(v interface{}) error {
	return gDBConn.AutoMigrate(v).Error
}

func InsertOne(v interface{}) error {
	return gDBConn.Create(v).Error
}

func Save(v interface{}) error {
	return gDBConn.Save(v).Error
}

func UpdateByColumn(conditions map[string]interface{}, v interface{}) error {
	db := gDBConn
	for col, val := range conditions {
		db = db.Where(fmt.Sprintf("%s = ?", col), val)
	}
	return db.Model(v).Updates(v).Error
}

func FindByID(id id.ID, v interface{}) (bool, error) {
	err := gDBConn.Where("id = ?", id).First(v).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func FindByColumn(colName string, colVal, v interface{}) (bool, error) {
	err := gDBConn.Where(map[string]interface{}{colName: colVal}).Find(v).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func FindOneByColumn(colName string, colVal, v interface{}) (bool, error) {
	err := gDBConn.Where(map[string]interface{}{colName: colVal}).First(v).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func FindOne(conditions map[string]interface{}, v interface{}) (bool, error) {
	db := gDBConn
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
	db := gDBConn
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
