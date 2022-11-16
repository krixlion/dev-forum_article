#!make
include .env
export $(shell sed 's/=.*//' .env)

mod-init: # param: mod_name
	go mod init	$(mod_name)
	go mod tidy
	go mod vendor

build:
	go build

run-dev:
	docker compose -f docker-compose.dev.yml up -d

run-local:
	go run cmd/main.go

push-image: # param: version
	docker build deployment -t krixlion/$(PROJECT_NAME)_$(AGGREGATE_ID):$(version)
	docker push krixlion/$(PROJECT_NAME)_$(AGGREGATE_ID):$(version)
