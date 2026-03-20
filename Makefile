.PHONY: proto build run test docker-up docker-down clean

proto:
	protoc \
		--proto_path=proto \
		--proto_path=/usr/include \
		--go_out=gen/shipment \
		--go_opt=paths=source_relative \
		--go-grpc_out=gen/shipment \
		--go-grpc_opt=paths=source_relative \
		proto/shipment.proto

build:
	go build -o shipment-service ./cmd/main.go

run: build
	./shipment-service

test:
	go test ./internal/... -v

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down -v

clean:
	rm -f shipment-service
