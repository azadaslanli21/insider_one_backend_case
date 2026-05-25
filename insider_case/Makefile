APP=football-simulator

.PHONY: run build test docker-up docker-down reset

run:
	go run ./cmd/api

build:
	go build -o bin/$(APP) ./cmd/api

test:
	go test ./...

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

reset:
	curl -X POST http://localhost:8080/league/reset
