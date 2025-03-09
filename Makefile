.PHONY: build clean test run 
 
build: 
	go build -o bin/go-initializer cmd/go-initializer/main.go 
 
test: 
	go test ./... 
 
clean: 
	rm -rf bin/ 
 
run: 
	go run cmd/go-initializer/main.go 
