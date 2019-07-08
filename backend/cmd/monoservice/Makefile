PROJECT_NAME=monoservice

# PROJECT DIR PATHS
PATH_BACKEND=../..
PATH_BACKEND_SRC=$(PATH_BACKEND)/src

## PROJECT SETUP
init: dev-db-setup

## PROJECT DEVELOP
all: build dev-run

build: dev-bin-build
	
dev-run: dev-db-stop dev-db-start dev-bin-run

## BINARY
BIN_NAME=$(PROJECT_NAME).bin
GO=go
GO_BUILD=$(GO) build

dev-bin-build:
	$(GO_BUILD) -o $(BIN_NAME) main.go 

dev-bin-run: dev-bin-build
	$(DEV_ENV_VARS) ./$(BIN_NAME)

# TEST - BACKEND CODE
dev-bin-test-run:
	$(DEV_ENV_TEST_VARS) $(GO) test -count=1 -p=1 $(PATH_BACKEND)/...

## DATABASE
DEV_DB_DIR=$(PATH_BACKEND)/data
DEV_DB_NAME=nfactorvault
DEV_DB_PORT=5432
DEV_DB_HOST=localhost
DEV_DB_ENV_VARS=POSTGRES_PORT=$(DEV_DB_PORT) POSTGRES_HOST=$(DEV_DB_HOST) POSTGRES_DBNAME=$(DEV_DB_NAME)


dev-db-setup:
	@echo "Installing Postgresql for macOS"
	-brew install postgresql
	-initdb -D $(DEV_DB_DIR) --no-locale --encoding=UTF8
	-mkdir $(DEV_DB_DIR)
	-pg_ctl -D $(DEV_DB_DIR) -l $(DEV_DB_DIR)/postgres.log start
	-createdb $(DB_NAME)
	pg_ctl -D $(DEV_DB_DIR) -l $(DEV_DB_DIR)/postgres.log stop

dev-db-start: dev-db-start-bg

dev-db-stop:
	-pg_ctl -D $(DEV_DB_DIR) -l $(DEV_DB_DIR)/postgres.log stop

dev-db-start-bg:
	pg_ctl -D $(DEV_DB_DIR) -l $(DEV_DB_DIR)/postgres.log start

dev-db-start-fg:
	postgres -D $(DEV_DB_DIR) -p $(DEV_DB_PORT) -h $(DEV_DB_HOST)

# TEST - DATABASE
### Note: Let's use the same DEV database with a different DB name
DEV_DB_TEST_DIR=$(DEV_DB_DIR)
DEV_DB_TEST_NAME=$(DEV_DB_NAME)_test
DEV_DB_TEST_ENV_VARS=POSTGRES_PORT=$(DEV_DB_PORT) POSTGRES_HOST=$(DEV_DB_HOST) POSTGRES_DBNAME=$(DEV_DB_TEST_NAME)

dev-db-test-setup:
	--mkdir $(DEV_DB_TEST_DIR)
	-initdb -D $(DEV_DB_TEST_DIR) --no-locale --encoding=UTF8
	-pg_ctl -D $(DEV_DB_TEST_DIR) -l $(DEV_DB_TEST_DIR)/postgres.log start
	-createdb $(DEV_DB_TEST_NAME)
	-pg_ctl -D $(DEV_DB_TEST_DIR) -l $(DEV_DB_TEST_DIR)/postgres.log stop

dev-db-test-start: dev-db-test-start-bg

dev-db-test-stop:
	-pg_ctl -D $(DEV_DB_TEST_DIR) -l $(DEV_DB_TEST_DIR)/postgres.log stop

dev-db-test-start-bg:
	pg_ctl -D $(DEV_DB_TEST_DIR) -l $(DEV_DB_TEST_DIR)/postgres.log start

dev-db-test-start-fg:
	postgres -D $(DEV_DB_TEST_DIR) -p $(DEV_DB_TEST_PORT) -h $(DEV_DB_TEST_HOST)

## ENV Variables
DEV_ENV_VARS=ENV=DEV $(DEV_DB_ENV_VARS)

DEV_ENV_TEST_VARS=ENV=TEST $(DEV_DB_TEST_ENV_VARS)
