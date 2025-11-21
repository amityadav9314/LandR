.PHONY: run build proto clean

# Variables
APP_NAME=landr-backend
PROTO_DIR=proto
GO_OUT_DIR=.
MODULE_NAME=github.com/amityadav/landr

run:
	go run cmd/server/main.go

build:
	go build -o bin/$(APP_NAME) cmd/server/main.go

proto:
	protoc --go_out=$(GO_OUT_DIR) --go_opt=module=$(MODULE_NAME) \
	--go-grpc_out=$(GO_OUT_DIR) --go-grpc_opt=module=$(MODULE_NAME) \
	$(PROTO_DIR)/*.proto

clean:
	rm -rf bin
