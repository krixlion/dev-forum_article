FROM golang:1.19.3-alpine

ENV GO111MODULE=on
ENV GOFLAGS=-mod=vendor
ENV APP_HOME /go/src/app

RUN mkdir -p "$APP_HOME"
WORKDIR "$APP_HOME"

ADD ../. .

RUN go mod tidy
RUN go mod vendor
RUN go build cmd/main.go
EXPOSE 50051
CMD [ "./main" ]