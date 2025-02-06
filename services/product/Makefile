LOCAL_BIN:=$(CURDIR)/.bin

run:
	@go build -o .bin/main cmd/app/main.go
	@.bin/main

run.pprof:
	@go build -o .bin/main cmd/app/main.go
	@go build -o .bin/pprof ./cmd/pprof/main.go
	@bash -c '.bin/main & .bin/pprof'

install-deps:
	@GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	@GOBIN=$(LOCAL_BIN) go install -mod=mod google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

get-deps:
	@go get -u google.golang.org/protobuf/cmd/protoc-gen-go
	@go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc

generate:
	@make generate-product-api-v1

generate-product-api-v1:
	@mkdir -p pkg/grpc/product/v1
	@protoc --proto_path api/grpc/product/v1 \
	--go_out=pkg/grpc/product/v1 --go_opt=paths=source_relative \
	--plugin=protoc-gen-go=.bin/protoc-gen-go \
	--go-grpc_out=pkg/grpc/product/v1 --go-grpc_opt=paths=source_relative \
	--plugin=protoc-gen-go-grpc=.bin/protoc-gen-go-grpc \
	api/grpc/product/v1/product.proto

generate-rsa256-keypair:
	@ssh-keygen -t rsa -b 4096 -m PEM -f jwt.key
	@openssl rsa -in jwt.key -pubout -outform PEM -out jwt.key.pub

include .env
init.db:
	@docker run --name=product-db -e POSTGRES_PASSWORD=${PG_PASSWORD} -p ${PG_PORT}:5432 -d postgres
	@sleep 2
	@make migrate.up

stop.db:
	@docker rm -f product-db

stop.db.test:
	@docker rm -f product-test-db

exec.db:
	@docker exec -it product-db bash -c "psql -U ${PG_USER} -d ${PG_DBNAME}"

migrate.up:
	@migrate -source file://migrations -database ${PG_MIGRATIONS_URL} up

migrate.down:
	@migrate -source file://migrations -database ${PG_MIGRATIONS_URL} down

generate.mocks.handlers:
	@mockgen -source=internal/application/interfaces/product.go -destination=internal/infrastructure/grpc/mocks/mocks.go

generate.mocks.service:
	@mockgen -source=internal/domain/repository/product.go -destination=internal/application/service/mocks/mocks.go

test.integration:
	@docker run --name=product-test-db -e POSTGRES_PASSWORD=${PG_PASSWORD} -p ${PG_TEST_PORT}:5432 -d postgres
	@bash -c 'trap "make stop.db.test" EXIT; \
		go test ./tests/integration -v; RESULT=$$?; \
		exit $$RESULT'