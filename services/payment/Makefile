include .env
init.db:
	@docker run --name=payment-db -e POSTGRES_PASSWORD=${PG_PASSWORD} -p ${PG_PORT}:5432 -d postgres
	@sleep 2
	@make migrate.up

stop.db:
	@docker rm -f payment-db

stop.db.test:
	@docker rm -f payment-test-db

exec.db:
	@docker exec -it payment-db bash -c "psql -U ${PG_USER} -d ${PG_DBNAME}"

migrate.up:
	@migrate -source file://migrations -database ${PG_MIGRATIONS_URL} up

migrate.down:
	@migrate -source file://migrations -database ${PG_MIGRATIONS_URL} down

generate.mocks.handlers:
	@mockgen -source=internal/application/interfaces/payment.go -destination=internal/interfaces/grpc_server/mocks/mocks.go

test.integration:
	@docker run --name=payment-test-db -e POSTGRES_PASSWORD=${PG_PASSWORD} -p ${PG_TEST_PORT}:5432 -d postgres
	@sleep 1
	@bash -c 'trap "docker rm -f payment-test-db" EXIT; \
		go test ./tests/integration -v; RESULT=$$?; \
		exit $$RESULT'

.PHONY: \
	init.db stop.db exec.db \
	migrate.up migrate.down