build:
	CGO_ENABLED=0 go build -o ./bin/cge-parser

run:
	go run .

test:
	go test ./...
 
clean:
	go clean
	rm ./bin/cge-parser
	rmdir ./bin
