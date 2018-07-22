default: test

.PHONY: test run compose.build compose.up compose.upb compose.redis

test:
	@REDIS_ADDR=localhost:6379 \
	GOCACHE=off \
	go test ./...

run:
	@CRAWLER_ROOT=$(PWD)/../crawler \
	REDIS_ADDR=localhost:6379 \
	go run main.go -logtostderr=true

compose.build:
	@docker-compose build

compose.up:
	@docker-compose up

compose.upb: compose.build compose.up

compose.redis:
	@docker-compose up redis
