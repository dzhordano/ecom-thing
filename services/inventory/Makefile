LOCAL_BIN:=$(CURDIR)/.bin

install-deps:
	@GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@GOBIN=$(LOCAL_BIN) go install -mod=mod google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

get-deps:
	@go get -u google.golang.org/protobuf/cmd/protoc-gen-go
	@go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc

generate-buf:
	@buf generate

generate:
	@make generate-inventory-api-v1

generate-inventory-api-v1:
	@mkdir -p pkg/api/inventory/v1
	@protoc --proto_path ../../proto/api/inventory/v1 \
	--go_out=pkg/api/inventory/v1 --go_opt=paths=source_relative \
	--plugin=protoc-gen-go=.bin/protoc-gen-go \
	--go-grpc_out=pkg/api/inventory/v1 --go-grpc_opt=paths=source_relative \
	--plugin=protoc-gen-go-grpc=.bin/protoc-gen-go-grpc \
	../../proto/api/inventory/v1/inventory.proto

generate-rsa256-keypair:
	@ssh-keygen -t rsa -b 4096 -m PEM -f jwt.key
	@openssl rsa -in jwt.key -pubout -outform PEM -out jwt.key.pub

include .env

run.nodocker:
	@make init.db
	@go build -o .bin/main cmd/app/main.go
	@.bin/main

migrate.up:
	@migrate -source file://migrations -database ${PG_MIGRATIONS_URL} up

init.db:
	@docker run --name=inventory-db -e POSTGRES_PASSWORD=${PG_PASSWORD} -p ${PG_PORT}:5432 -d postgres

exec.db:
	@docker exec -it inventory-db bash -c "psql -U ${PG_USER} -d ${PG_DBNAME}"

test.load:
	@ghz --insecure --proto proto/api/inventory/v1/inventory.proto --call api.inventory.v1.InventoryService/AddQuantity \
	--rps 100 --total 3000 \
	-d '{"id": "00000000-0000-0000-0000-000000000000", "quantity": "1"}' \
	localhost:50001