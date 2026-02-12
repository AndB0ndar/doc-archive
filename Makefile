migrate_file ?= init_schema
DATABASE_URL ?= "postgres://user:pass@localhost:5432/docdb?sslmode=disable"


.PHONY: up migrate-create migrate-up migrate-down run


up:
	docker-compose up -d postgres

migrate-create:
	migrate create -ext sql -dir migrations -seq $(migrate_file)

migrate-up:
	migrate -path migrations -database $(DATABASE_URL) up

migrate-down:
	migrate -path migrations -database $(DATABASE_URL) down

run:
	go run cmd/api/main.go

