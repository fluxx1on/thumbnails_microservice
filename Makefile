PHONY: generate
generate:
	mkdir -p grpc
	protoc --go_out=grpc --go_opt=paths=import \
	--go-grpc_out=grpc --go-grpc_opt=paths=import \
	api/thumbnails.proto
	mv grpc/github.com/fluxx1on/thumbnails_microservice/internal/grpc/proto* internal/grpc/proto/
	rm -rf grpc/github.com

set_environment:
	echo "ROOT_DIR=\"$(pwd)/\"" >> .env
	echo "MEDIA_DIR=\"$(pwd)/media/\"" >> .env
	echo "LOG_FILE=\"$(pwd)/service_log.log\"" >> .env

setup: 
	go mod tidy
	make generate
	make set_environment

test:
	go test -v -count=1  ./...

.PHONY: cover
cover:
	go test -short -count=1 -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm coverage.out

build:
	mkdir -p bin
	go build -C cmd -o ../bin/server

run: build
	./bin/server