.PHONY: docker-run docker-run-db docker-run-monoservice docker-run-go-test


# Docker Run
docker-run: docker-run-db docker-run-api
docker-run-test: docker-run-db docker-run-go-test 
docker-run-db:
	docker-compose up --build --remove-orphans -d db
docker-run-api: 
	docker-compose up --build --remove-orphans monoservice
docker-run-go-test:
	docker-compose up --build --remove-orphans go-test
docker-stop:
	docker-compose stop db monoservice go-test