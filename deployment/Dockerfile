FROM golang:1.22 AS base
WORKDIR /go/src/dev_forum-article
ENV GO111MODULE=on
ENV GOFLAGS=-mod=vendor
# Make sure you run `docker build` from the project root or set the context accordingly.
ADD . .
# Make a user to import to scratch image
RUN useradd -u 10001 -m app
RUN go mod tidy && go mod vendor 

FROM base as build
# Must be set in order to avoid dynamic links to libraries unavailable in alpine/scratch images.
ENV CGO_ENABLED=0
RUN go build -o main cmd/main.go

FROM base as test
# Must be set in order to enable `go -race` flag.
ENV CGO_ENABLED=1
WORKDIR /app
USER app
COPY --from=build --chown=app:app /go/src/dev_forum-article/ /app/
RUN go build -o main cmd/main.go
ENTRYPOINT [ "/app/main" ]

FROM scratch as prod
WORKDIR /app
USER app

COPY --from=build  /go/src/dev_forum-article/main /app/main
COPY --from=build /go/src/dev_forum-article/.env /app/.env
# A user is required to allow Go os/user calls without CGO enabled.
COPY --from=build /etc/passwd /etc/passwd

EXPOSE 50051
EXPOSE 2223

ENTRYPOINT [ "/app/main" ]
