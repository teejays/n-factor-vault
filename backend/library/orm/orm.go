package orm

import (
	"fmt"
	"reflect"
	"time"

	// github.com/jinzhu/gorm/dialects/postgres is needed to connect gorm to a Postgres database
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/jinzhu/gorm"
	"github.com/teejays/clog"

	"github.com/teejays/n-factor-vault/backend/library/env"
)

// TODO: type DB {*gorm.DB} would just allows us to add more wrapper functions
// over *gorm.DB

var gDB *gorm.DB

// Init initializes the ORM package by connecting to the database. Init needs to be run before the package can be
// properly used. This expects you to have set the environment variables POSTGRES_PORT, POSTGRES_HOST and POSTGRES_DBNAME.
func Init() error {
	// Get the connection string that we can use to connect to the database server
	connStr, err := getPostgresConnectionString()
	if err != nil {
		return fmt.Errorf("Could not connect get postgres connection string: %v", err)
	}

	// If we can't connect to DB, we should probably try a few times with somewait
	// Sometimes, the DB isn't up and ready yet
	var db *gorm.DB
	var retryAttempts = 5
	for i := 0; i < retryAttempts; i++ {
		db, err = gorm.Open("postgres", connStr)
		if err != nil {
			message := fmt.Sprintf("orm: Could not connect to database (attempt #%d out of %d): %v", i+1, retryAttempts, err)
			if i == retryAttempts-1 {
				return fmt.Errorf(message)
			}
			clog.Error(message)
			time.Sleep(3 * time.Second)
			continue
		}

		break
	}

	// TODO: Inject clog as the default logger for gorm
	// db.SetLogger(gorm.Logger{})

	// By default, set log mode to false
	db.LogMode(false)

	// For DEV environment or if env var LOG_ORM is set to true/1, log more ORM stuff
	if env.GetEnv() == env.DEV || env.GetBoolOrDefault("LOG_ORM", false) {
		db.LogMode(true)
	}

	gDB = db
	clog.Infof("orm: DB connection opened: %+v", gDB)
	return nil
}

// Close closes the orm connection
func Close() {
	if gDB != nil {
		return
	}
	gDB.Close()
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

	// Get the user and password
	user, _ := env.GetEnvVar("POSTGRES_USER")
	password, _ := env.GetEnvVar("POSTGRES_PWD")

	var str = "host=%s port=%d dbname=%s sslmode=disable"
	var args = []interface{}{host, port, dbName}
	if user != "" {
		str += " user=%s"
		args = append(args, user)
	}
	if password != "" {
		str += " password=%s"
		args = append(args, password)
	}
	return fmt.Sprintf(str, args...), nil
}

// RegisterModel registers the provided struct as a gorm model, creating the database table
// in the process. All the magic stuff that goes along with creating a model should happen here.
//
// It does a few things:
//
// Create the main table for the struct
// Handle the migration if the struct is being changed
// TODO: Create a history table that keeps historic rows of the main table
// Create a trigger that inserts data into the history table if a row is mutated in the main table
func RegisterModel(v Entity) error {
	clog.Infof("orm: Registering model %T", v)

	// Create/Migrate the main table
	db := gDB.AutoMigrate(v)
	if db.Error != nil {
		return db.Error
	}

	// TODO: Create a history table

	return gDB.AutoMigrate(v).Error
}

// RegisterModels register's multiple models in one go.
func RegisterModels(models ...Entity) error {
	for _, v := range models {
		if err := RegisterModel(v); err != nil {
			return fmt.Errorf("registering %s", reflect.TypeOf(v))
		}
	}
	return nil
}
