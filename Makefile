PHONY: generate
generate:
	mkdir -p internal
	protoc --go_out=internal --go_opt=paths=import \
	--go-grpc_out=internal --go-grpc_opt=paths=import \
	api/thumbnails.proto
	mv grpc/github.com/fluxx1on/thumbnails_microservice/internal/grpc/proto* internal/grpc/proto/
	rm -rf grpc/github.com

set_environment:
	touch .env
	echo "ROOT_DIR=\"$(pwd)/\"" >> .env
	echo "MEDIA_DIR=\"$(pwd)/media/\"" >> .env
	echo "LOG_FILE=\"$(pwd)/service_log.log\"" >> .env
	echo "STAGE=\"dev\"" >> .env
	echo "SERVER_ADDRESS=\"127.0.0.1:50051\"" >> .env
	echo "LISTENER_PROTOCOL=\"tcp\"" >> .env
	echo "REDIS_ADDRESS=\":6379\"" >> .env
	echo "REDIS_CONNECTION_POOL=\"10\"" >> .env
	echo "REDIS_DB=\"0\"" >> .env
	echo "YOUTUBE_APIKEY=" >> .env

setup: 
	go mod tidy
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