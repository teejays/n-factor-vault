PROJECT_NAME=monoservice


## PROJECT SETUP
init: dev-db-setup

## PROJECT DEVELOP
all: build dev-run

build: dev-bin-build
	
dev-run: dev-db-start dev-bin-run

## BINARY
BIN_NAME=$(PROJECT_NAME).bin
GO=go
GO_BUILD=$(GO) build

dev-bin-build:
	$(GO_BUILD) -o $(BIN_NAME) main.go 

dev-bin-run: dev-bin-build
	./$(BIN_NAME)

## DATABASE
DEV_DB_DIR=../../data
DB_NAME=nfactorvault

dev-db-start: dev-db-start-bg

dev-db-stop:
	pg_ctl -D $(DEV_DB_DIR) -l $(DEV_DB_DIR)/postgres.log stop


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
	postgres -D $(DEV_DB_DIR)

