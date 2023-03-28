.PHONY: build protobuf run test clean

build: protobuf
	CGO_ENABLED=0 go build -o ./bin/cge-parser

protobuf:
	protoc -I=protobuf/ --go_out=. protobuf/schema.proto

run: protobuf
	go run .

test: protobuf
	go test ./...
 
clean:
	go clean
	rm ./bin/cge-parser
	rmdir ./bin
