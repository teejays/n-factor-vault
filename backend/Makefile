.PHONY: run run-db run-monoservice run-go-test

run: run-db run-api
run-test: run-db run-go-test 
run-db:
	docker-compose up --build --remove-orphans -d db
run-api: 
	docker-compose up --build --remove-orphans monoservice
run-go-test:
	docker-compose up --build --remove-orphans go-test
stop:
	docker-compose kill -s SIGINT db monoservice go-test