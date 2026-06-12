.PHONY: build test run clean up logs down restart health build-docker

# 本地开发
build:
	mkdir -p bin
	go build -ldflags="-s -w" -o bin/vision-assistant.exe ./cmd/server

test:
	go test ./... -v -count=1

run: build
	./bin/vision-assistant.exe

clean:
	rm -rf bin/

# Docker (WSL / Linux)
build-docker:
	docker compose build

up:
	docker compose --env-file .env up -d

logs:
	docker compose logs -f

down:
	docker compose down -v

restart: down build-docker up

health:
	curl -s http://localhost:8080/health
