PROJECT_NAME=monoservice

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
	$(ENV_VARS) ./$(BIN_NAME)

## DATABASE
DEV_DB_DIR=../../data
DEV_DB_NAME=nfactorvault
DEV_DB_PORT=5432
DEV_DB_HOST=localhost

DB_ENV_VARS=POSTGRES_PORT=$(DEV_DB_PORT) POSTGRES_HOST=$(DEV_DB_HOST) POSTGRES_DBNAME=$(DEV_DB_NAME)

dev-db-start: dev-db-start-bg

dev-db-stop:
	-pg_ctl -D $(DEV_DB_DIR) -l $(DEV_DB_DIR)/postgres.log stop


dev-db-setup:
	@echo "Installing Postgresql for macOS"
	-brew install postgresql
	-initdb -D $(DEV_DB_DIR) --no-locale --encoding=UTF8
	pg_ctl -D $(DEV_DB_DIR) -l $(DEV_DB_DIR)/postgres.log start
	-createdb $(DB_NAME)
	pg_ctl -D $(DEV_DB_DIR) -l $(DEV_DB_DIR)/postgres.log stop

dev-db-start-bg:
	pg_ctl -D $(DEV_DB_DIR) -l $(DEV_DB_DIR)/postgres.log start

dev-db-start-fg:
	postgres -D $(DEV_DB_DIR) -p $(DEV_DB_PORT) -h $(DEV_DB_HOST)


## ENV Variables
ENV_VARS=ENV=DEV $(DB_ENV_VARS)
