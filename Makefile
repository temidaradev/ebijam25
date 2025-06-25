run:
	@go run cmd/main.go
build:
	@go build -o bin/cli cmd/main.go
wasm:
	@GOOS=js GOARCH=wasm go build -o dist/wasm/main.wasm cmd/main.go
