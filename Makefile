build:
	go build -o bin/bank

run: build
	./bin/bank

test:
	go test -v ./...