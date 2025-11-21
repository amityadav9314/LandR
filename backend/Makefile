.PHONY: run build proto clean

# Variables
APP_NAME=landr-backend
PROTO_DIR=proto
GO_OUT_DIR=.

run:
	go run cmd/server/main.go

build:
	go build -o bin/$(APP_NAME) cmd/server/main.go

proto:
	protoc --go_out=$(GO_OUT_DIR) --go_opt=paths=source_relative \
	--go-grpc_out=$(GO_OUT_DIR) --go-grpc_opt=paths=source_relative \
	$(PROTO_DIR)/*.proto

clean:
	rm -rf bin
