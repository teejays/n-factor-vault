.PHONY: docker-run docker-run-db docker-run-monoservice docker-run-go-test


# Docker Run
docker-run: docker-run-db docker-run-api
docker-run-test: docker-run-test-db docker-run-go-test docker-stop-test-db

docker-run-db:
	docker-compose up --build --remove-orphans -d db
docker-run-api: 
	docker-compose up --build --remove-orphans monoservice

docker-run-test-db:
	docker-compose up --build --remove-orphans -d test_db
docker-stop-test-db:
	docker-compose stop test_db
docker-run-go-test:
	docker-compose up --build --remove-orphans go_test

docker-stop-all:
	docker-compose stop