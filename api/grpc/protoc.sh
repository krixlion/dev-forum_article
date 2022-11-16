#! /bin/sh.
DIR="$(dirname "${BASH_SOURCE[0]}")"
DIR="$(realpath "${DIR}")"
GO_PB_PATH="$(cd $DIR/pkg/grpc/pb && pwd)"

protoc --go_out=paths=source_relative:$GO_PB_PATH --doc_out=$DIR/docs --doc_opt=markdown,docs.md --go-grpc_out=paths=source_relative:$GO_PB_PATH -I $DIR/api/grpc article-service.proto
protoc-go-inject-tag -input="$GO_PB_PATH/*.pb.go"