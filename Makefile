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

build:
	mkdir -p bin
	go build -C cmd -o ../bin/server
	./bin/server

run:
	make build