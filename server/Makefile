.PHONY: up migrate-create migrate-up migrate-down run

DATABASE_URL ?= "postgres://user:pass@localhost:5432/docdb?sslmode=disable"


migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

migrate-up:
	migrate -path migrations -database $(DATABASE_URL) up

migrate-down:
	migrate -path migrations -database $(DATABASE_URL) down

swagger:
	swag init -g cmd/server/main.go
