# dev_forum-article

Article-service is responsible of storing and operating on articles that users create in dev_forum system.

It's dependent on:
  - [Redis](https://redis.io/docs/get-started) for query storage.
  - [EventStore](https://www.cockroachlabs.com/docs/cockroachcloud/quickstart) for storing articles and their update history.
  - [RabbitMQ](https://www.rabbitmq.com/#getstarted) for asynchronous communication with other components in the domain.
  - [OtelCollector](https://opentelemetry.io/docs/collector) for receiving and forwarding telemetry data.

## Set up

To set up the service copy `.env.example`, rename the file to `.env` and fill in any blank values. Make sure to place it at the same directory the executable and `Makefile` is.

### Using Go command
You need working [Go environment](https://go.dev/doc/install).
```
go mod vendor

go build cmd/main.go
```

### Using Docker
You need working [Docker environment](https://docs.docker.com/get-started).

You can build the service with the Dockerfile located in `deployment/Dockerfile`.

### Using Kubernetes
You need working [Kubernetes environment](https://kubernetes.io/docs/setup).

Kubernetes resources are defined in `deployment/k8s` and deployed using [Kustomize](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/).

Currently there are `stage` and `dev` overlays available and include any needed resources and configs.

You can deploy `dev` using `make`.
```
make k8s-run-dev

# To delete
make k8s-stop-dev
```

## Testing

Run unit and integration tests using Go command. Add `-short` flag to skip integration tests.
Make sure to set current working directory to project root.
```
go test ./... -race
```

Generate coverage report using `go tool cover`.
```
go test -coverprofile  cover.out ./...
go tool cover -html cover.out -o cover.html
```

If the service is deployed on kubernetes you can use `make`.
```
make k8s-integration-test

# or

make k8s-unit-test
```

## API
Service is exposing [gRPC](https://grpc.io/docs/what-is-grpc/introduction) API.

Regenerate `pb` packages after making changes to any of the `.proto` files located in `api/`.
You can use [go-grpc-gen](https://github.com/krixlion/go-grpc-gen) tool with `make grpc-gen`.
