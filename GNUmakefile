default: testacc

.PHONY: build
build:
	go build -v ./...

.PHONY: install
install: build
	go install -v ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: generate
generate:
	go generate ./...

.PHONY: fmt
fmt:
	gofmt -s -w .

.PHONY: test
test:
	go test -v -count=1 -parallel=10 ./...

.PHONY: testacc
testacc:
	TF_ACC=1 go test -v -count=1 -parallel=10 -timeout 120m ./...
