#!make
include .env
export $(shell sed 's/=.*//' .env)

mod-init: 
	go mod init	github.com/krixlion/$(PROJECT_NAME)_$(AGGREGATE_ID)
	go mod tidy
	go mod vendor

build:
	go build

run-dev:
	docker compose -f docker-compose.dev.yml --env-file .env build
	docker compose -f docker-compose.dev.yml --env-file .env up -d --remove-orphans

run-local:
	go run cmd/main.go

test:
	docker compose exec service go test -race ./...

push-image: # param: version
	docker build deployment/ -t krixlion/$(PROJECT_NAME)_$(AGGREGATE_ID):$(version)
	docker push krixlion/$(PROJECT_NAME)_$(AGGREGATE_ID):$(version)
