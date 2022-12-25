#!make
include .env
export $(shell sed 's/=.*//' .env)

docker-compose = docker compose -f docker-compose.dev.yml --env-file .env
k8s = kubectl -n dev

mod-init:
	go mod init	github.com/krixlion/$(PROJECT_NAME)_$(AGGREGATE_ID)
	go mod tidy
	go mod vendor

run-local:
	go run cmd/main.go

build-local:
	go build

push-image: # param: version
	docker build deployment/ -t krixlion/$(PROJECT_NAME)_$(AGGREGATE_ID):$(version)
	docker push krixlion/$(PROJECT_NAME)_$(AGGREGATE_ID):$(version)


# ------------- Docker Compose -------------

docker-run-dev: #param: args
	$(docker-compose) build ${args}
	$(docker-compose) up -d --remove-orphans

docker-test: # param: args
	$(docker-compose) exec service go test -race ${args} ./...  

docker-test-gen-coverage:
	$(docker-compose) exec service go test -coverprofile cover.out ./...
	$(docker-compose) exec service go tool cover -html cover.out -o cover.html


# ------------- Kubernetes -------------

k8s-test: # param: args
	$(k8s) exec -it deploy/article-d -- go test -race ${args} ./...  

k8s-run-dev:
	- $(k8s) delete -R -f k8s/resources
	$(k8s) apply -R -f k8s/resources

k8s-setup-tools:
	kubectl apply -f k8s/dev-namespace.yml
	kubectl apply -R -f k8s/kubernetes-dashboard.yml
	kubectl apply -R -f k8s/metrics-server.yml

k8s-setup-telemetry:
	$(k8s) apply -R -f k8s/instrumentation/