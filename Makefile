run:
	docker-compose -f docker-compose.yml up

gqlgen:
	go run github.com/99designs/gqlgen generate

wire:
	go get github.com/google/wire/cmd/wire
	go run github.com/google/wire/cmd/wire ./internal/di/wire.go