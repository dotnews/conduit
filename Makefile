default: test

.PHONY: test redis run compose compose.up

test:
	@REDIS_ADDR=localhost:6379 GOCACHE=off go test ./...

redis:
	@docker-compose up redis

run:
	@REDIS_ADDR=localhost:6379 go run main.go -logtostderr=true

compose:
	@docker-compose build

compose.up: compose
	@docker-compose up
