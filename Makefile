up:
	docker-compose up --build -d

down:
	docker-compose down

proto:
	protoc --go_out=. --go-grpc_out=. proto/cache.proto

bench:
	curl "http://localhost:8080/benchmark?keys=1000"
	curl "http://localhost:8080/benchmark?keys=10000"
	curl "http://localhost:8080/benchmark?keys=100000"

test:
	go test ./pkg/...

integration-test:
	TEST_NODE_ADDR=localhost:50051 go test -tags=integration ./internal/gateway/...

test-all: test integration-test

lint:
	golangci-lint run ./...

security:
	gosec ./...