#! /bin/sh.
DIR="$(dirname "${BASH_SOURCE[0]}")"
DIR="$(realpath "${DIR}")"
GO_PB_PATH="$(cd $DIR/../../internal/pb/ && pwd)"

echo This script is not adjusted to this service, change the protofile before executing this script.
echo "*Aborting*"

exit
protoc --go_out=paths=source_relative:$GO_PB_PATH --go-grpc_out=paths=source_relative:$GO_PB_PATH -I $DIR *.proto
protoc-go-inject-tag -input="$GO_PB_PATH/*.pb.go"