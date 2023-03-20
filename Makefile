.PHONY: deps
deps:
	go mod tidy
	go mod vendor

.PHONY: mocks
mocks:
	go generate