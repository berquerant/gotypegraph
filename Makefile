.PHONY: build
build:
	@mkdir -p dist
	@go build -o dist/gotypegraph ./main.go

.PHONY: test
test:
	@go test ./...
