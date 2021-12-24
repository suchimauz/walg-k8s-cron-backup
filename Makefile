.PHONY:
.SILENT:
.DEFAULT_GOAL := run

run:
	go run cmd/app/main.go

environment:
	cp env.dist .env

test:
	go test --short -coverprofile=cover.out -v ./...
	make test.coverage

test.coverage:
	go tool cover -func=cover.out