include .env
export $(shell sed 's/=.*//' .env)

MIGRATIONS_PATH := internal/adapter/database/migrations
INIT_PATH := internal/adapter/database/init

build:
	@go generate ./...
	@go build -o ./bin/go-fiber-template ./main.go
start:
	@go run main.go service
lint:
	@golangci-lint run
tests:
	@go test -v ./...
tests-%:
	@go test -v ./... -run=$(shell echo $* | sed 's/_/./g')
testsum:
	@gotestsum --format testname
swagger:
	@swag init
migration-%:
	@migrate create -ext sql -dir $(MIGRATIONS_PATH) create-table-$(subst :,_,$*)
migrate-up:
	@migrate -database "postgres://$(DATABASE_USER):$(DATABASE_PASSWORD)@$(DATABASE_HOST):5432/$(DATABASE_NAME)?sslmode=disable" -path $(MIGRATIONS_PATH) up
migrate-down:
	@migrate -database "postgres://$(DATABASE_USER):$(DATABASE_PASSWORD)@$(DATABASE_HOST):5432/$(DATABASE_NAME)?sslmode=disable" -path $(MIGRATIONS_PATH) down
migrate-docker-up:
	@docker run -v ./$(MIGRATIONS_PATH):/migrations --network go-fiber-template_go-network migrate/migrate -path=/migrations/ -database postgres://$(DATABASE_USER):$(DATABASE_PASSWORD)@$(DATABASE_HOST):5432/$(DATABASE_NAME)?sslmode=disable up
migrate-docker-down:
	@docker run -v ./$(MIGRATIONS_PATH):/migrations --network go-fiber-template_go-network migrate/migrate -path=/migrations/ -database postgres://$(DATABASE_USER):$(DATABASE_PASSWORD)@$(DATABASE_HOST):5432/$(DATABASE_NAME)?sslmode=disable down -all
docker:
	@chmod -R 755 $(INIT_PATH)
	@docker compose up --build
docker-test:
	@docker compose up -d && make tests
docker-down:
	@docker compose down --rmi all --volumes --remove-orphans
docker-cache:
	@docker builder prune -f