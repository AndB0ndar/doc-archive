.PHONY: up migrate run

up:
	docker-compose up -d

migrate:
	migrate -path migrations -database "postgres://user:pass@localhost:5432/docdb?sslmode=disable" up

run:
	go run cmd/api/main.go
