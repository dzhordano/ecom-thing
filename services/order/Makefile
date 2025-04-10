include .env
init.db:
	@docker run --name=order-db -e POSTGRES_PASSWORD=${PG_PASSWORD} -p ${PG_PORT}:5432 -d postgres
	@sleep 2
	@make migrate.up

stop.db:
	@docker rm -f order-db

stop.db.test:
	@docker rm -f order-test-db

exec.db:
	@docker exec -it order-db bash -c "psql -U ${PG_USER} -d ${PG_DBNAME}"

migrate.up:
	@migrate -source file://migrations -database ${PG_MIGRATIONS_URL} up

migrate.down:
	@migrate -source file://migrations -database ${PG_MIGRATIONS_URL} down

test.load:
	@ghz --insecure --proto proto/api/order/v1/order.proto --call api.order.v1.OrderService/GetOrder \
	--rps 1000 --total 5000 \
	-d '{"order_id": "00000000-0000-0000-0000-000000000001"}' \
	localhost:50003

generate.mocks.handlers:
	@mockgen -source=internal/application/interfaces/order.go -destination=internal/interfaces/grpc_server/mocks/mocks.go

.PHONY: init.db stop.db exec.db migrate.up migrate.down