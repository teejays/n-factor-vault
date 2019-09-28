.PHONY: docker-run docker-run-db docker-run-monoservice docker-run-go-test


# Docker Run
docker-run: docker-run-db docker-run-api
docker-run-test: docker-run-test-db docker-run-go-test docker-stop-test-db

# Main Build Commands
docker-run-db:
	docker-compose up --build --remove-orphans -d db
docker-run-api: 
	docker-compose up --build --remove-orphans monoservice

# Build Go locally for fast usage
local-go-build:
	go build -o monoservice.bin cmd/monoservice/main.go

# Test Commands
docker-run-test-db:
	docker-compose up --build --remove-orphans -d test_db
docker-stop-test-db:
	docker-compose stop test_db
docker-run-go-test:
	docker-compose up --build --remove-orphans go_test

test-db-connect:
	PGPASSWORD=docker psql -h localhost -p 8432 -U docker -w -d nfactorvault

## Kill Commands
docker-stop-all:
	docker-compose stop