# Status
🚧 **Under Development** 🚧

This repository is a part of an ongoing project and is currently under active development. I'm continuously working on adding features, fixing bugs, and improving documentation. 
Although this is a one-man project, contributions are welcome.
Please feel free to open issues or submit pull requests.

# dev_forum-article
[![GoDoc](https://godoc.org/github.com/krixlion/dev_forum-article?status.svg)](https://godoc.org/github.com/krixlion/dev_forum-article)
[![Coverage Status](https://coveralls.io/repos/github/krixlion/dev_forum-article/badge.svg?branch=dev)](https://coveralls.io/github/krixlion/dev_forum-article?branch=dev)
[![Go Report Card](https://goreportcard.com/badge/github.com/krixlion/dev_forum-article)](https://goreportcard.com/report/github.com/krixlion/dev_forum-article)
[![GitHub License](https://img.shields.io/github/license/krixlion/dev_forum-article)](LICENSE)

Article-service is responsible of storing and operating on articles that users create in dev_forum system.

It's dependent on:
  - [Redis](https://redis.io/docs/get-started) for query storage.
  - [EventStore](https://www.cockroachlabs.com/docs/cockroachcloud/quickstart) for storing articles and their update history.
  - [RabbitMQ](https://www.rabbitmq.com/#getstarted) for asynchronous communication with other components in the domain.
  - [OtelCollector](https://opentelemetry.io/docs/collector) for receiving and forwarding telemetry data.

## Set up
Rename `.env.example` to `.env` and fill in missing values.


### Using Go command
You need working [Go environment](https://go.dev/doc/install).
```shell
go mod tidy
go mod vendor
go build cmd/main.go
```

### On Docker
You need a working [Docker environment](https://docs.docker.com/engine).

You can use the Dockerfile located in `deployment/` to build and run the service on a docker container.

```shell
make build-image version=latest 
``` 

```shell
docker run -p 50051:50051 krixlion/dev_forum-article:latest
```

### On Kubernetes (recommended)
You need a working [Kubernetes environment](https://kubernetes.io/docs/setup).

Kubernetes resources are defined in `deployment/k8s` and deployed using [Kustomize](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/).

Currently there are `stage`, `dev` and `test` overlays available and include any needed resources and configs.

Use `make` to apply manifests for dev_forum-article and needed DBs for either dev or stage environment.
Every `make` rule that depends on k8s accepts an `overlay` param which indicates the namespace for the rule.
```shell
make k8s-run overlay=<dev/stage/...>
```
```shell
# To delete
make k8s-stop overlay=<dev/stage/...>
```

## Testing
Run unit and integration tests using Go command.
Make sure to set current working directory to project root.
```shell
# Add `-short` flag to skip integration tests.
go test ./... -race
```

Generate coverage report using `go tool cover`.
```shell
go test -coverprofile  cover.out ./...
go tool cover -html cover.out -o cover.html
```

If the service is deployed on kubernetes you can use `make`.
```shell
make k8s-integration-test overlay=<dev/stage/...>
```
or
```shell
make k8s-unit-test overlay=<dev/stage/...>
```

## Documentation
For in-detail documentation refer to the [Wiki](https://github.com/krixlion/dev_forum-article/wiki).

## API
Service is exposing a [gRPC](https://grpc.io/docs/what-is-grpc/introduction) API.

Regenerate `pb` packages after making changes to any of the `.proto` files located in `api/`.
You can use [go-grpc-gen](https://github.com/krixlion/go-grpc-gen), a containerized tool for generating gRPC bindings, with `make grpc-gen`.
