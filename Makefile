#!make
include .env
export $(shell sed 's/=.*//' .env)

kubernetes = kubectl -n dev
docker-compose = docker compose -f docker-compose.dev.yml --env-file .env

mod-init:
	go mod init	github.com/krixlion/$(PROJECT_NAME)_$(AGGREGATE_ID)
	go mod tidy
	go mod vendor

push-image: # param: version
	docker build deployment/ -t krixlion/$(PROJECT_NAME)_$(AGGREGATE_ID):$(version)
	docker push krixlion/$(PROJECT_NAME)_$(AGGREGATE_ID):$(version)


# ------------- Kubernetes -------------

k8s-mount-project:
	mkdir -p /mnt/wsl/k8s-mount/${AGGREGATE_ID} && sudo mount --bind . /mnt/wsl/k8s-mount/${AGGREGATE_ID}

k8s-unit-test: # param: args
	$(kubernetes) exec -it deploy/article-d -- go test -short -race ${args} ./...  

k8s-integration-test: # param: args
	$(kubernetes) exec -it deploy/article-d -- go test -race ${args} ./...  

k8s-test-gen-coverage:
	$(kubernetes) exec -it deploy/article-d -- go test -coverprofile  cover.out ./...
	$(kubernetes) exec -it deploy/article-d -- go tool cover -html cover.out -o cover.html

k8s-run-dev:
	- $(kubernetes) delete -R -f deployment/k8s/dev/resources/
	$(kubernetes) apply -R -f deployment/k8s/dev/resources/

k8s-stop-dev:
	- $(kubernetes) delete -R -f deployment/k8s/dev/resources/