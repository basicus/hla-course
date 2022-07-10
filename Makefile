#!/usr/bin/make

DOCKER_BIN = $(shell command -v docker 2> /dev/null)

run: # запуск
	 go run ./cmd/service.go

generate_users:
	go run ./tests/generate-users/generator.go

test: # Прогон всех тестов
	 go test ./...

docker: # Сборка образа
	$(DOCKER_BIN) build --network=host -t service-sn --no-cache .

docker_clear: # Сборка мусорных образов
	yes | $(DOCKER_BIN) image prune

linter:
	golangci-lint run
