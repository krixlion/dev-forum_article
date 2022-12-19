#!make
include .env
export $(shell sed 's/=.*//' .env)

docker-compose = docker compose -f docker-compose.dev.yml --env-file .env

mod-init: 
	go mod init	github.com/krixlion/$(PROJECT_NAME)_$(AGGREGATE_ID)
	go mod tidy
	go mod vendor

build:
	go build

run-dev: #param: args
	$(docker-compose) build ${args}
	$(docker-compose) up -d --remove-orphans

run-local:
	go run cmd/main.go

test: # param: args
	$(docker-compose) exec service go test -race ${args} ./...  

test-gen-coverage:
	$(docker-compose) exec service go test -coverprofile cover.out ./...
	$(docker-compose) exec service go tool cover -html cover.out -o cover.html
	

push-image: # param: version
	docker build deployment/ -t krixlion/$(PROJECT_NAME)_$(AGGREGATE_ID):$(version)
	docker push krixlion/$(PROJECT_NAME)_$(AGGREGATE_ID):$(version)
