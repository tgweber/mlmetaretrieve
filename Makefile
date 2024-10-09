.PHONY: build
build:
	go build -o retrieve ./cmd/main.go 
build-debug:
	go build -gcflags "all=-N -l" -o retrieve ./cmd/main.go
run:
	go run ./cmd/main.go ./config.json
test:
	go test -v -covermode count -coverprofile cover.out ./...
