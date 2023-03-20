.PHONY: deps
deps:
	go mod tidy
	go mod vendor

.PHONY: mocks
mocks:
	go generate

.PHONY: test
test:
	go test ./... -count=1