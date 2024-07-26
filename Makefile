#!make
include .env
export $(shell sed 's/=.*//' .env)

overlay ?= dev
kubernetes = kubectl -n ${overlay}
docker-compose = docker compose -f docker-compose.dev.yml --env-file .env
overlays-path = deployment/k8s/overlays

mod-init:
	go mod init	github.com/krixlion/$(PROJECT_NAME)_$(AGGREGATE_ID)
	go mod tidy
	go mod vendor

grpc-gen:
	docker run --rm -v $(shell pwd):/app --env-file .env krixlion/go-grpc-gen:${GO_VERSION}

build-image: # params: version, args
	# OTEL_EXPORTER_OTLP_ENDPOINT variable imported from '.env' file triggers tracing in docker and causes issues.
	OTEL_EXPORTER_OTLP_ENDPOINT="" docker build . -f deployment/Dockerfile -t krixlion/$(PROJECT_NAME)-$(AGGREGATE_ID):$(version) ${args}
	
push-image: build-image # param: version
	docker push krixlion/$(PROJECT_NAME)-$(AGGREGATE_ID):$(version)

# ------------- Kubernetes -------------

k8s-mount-project:
	mkdir -p /mnt/wsl/k8s-mount/${AGGREGATE_ID} && sudo mount --bind $(shell pwd) /mnt/wsl/k8s-mount/${AGGREGATE_ID}

k8s-unit-test: # param: args
	$(kubernetes) exec -it deploy/${AGGREGATE_ID}-d -- go test -short -race ${args} ./...  

k8s-integration-test: # param: args
	$(kubernetes) exec -it deploy/${AGGREGATE_ID}-d -- go test -race ${args} ./...  

k8s-test-gen-coverage:
	$(kubernetes) exec -it deploy/${AGGREGATE_ID}-d -- go test -coverprofile cover.out ${args} ./...
	$(kubernetes) exec -it deploy/${AGGREGATE_ID}-d -- go tool cover -html cover.out -o cover.html

k8s-run: k8s-stop # param: overlay
	$(kubernetes) -k $(overlays-path)/${overlay} apply

k8s-stop: # param: overlay
	- $(kubernetes) -k $(overlays-path)/${overlay} delete 
