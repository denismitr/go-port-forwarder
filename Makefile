MASTER_URL := $(shell cat ~/.kube/config | grep 127 | awk -F 'server: ' '/server: /{print $$2}')
KUBE_CONFIG := $(shell cat ~/.kube/config | base64)

.PHONY: vars
vars:
	@echo master URL - ${MASTER_URL}

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

.PHONY: initialize
initialize: export K8S_MASTER_URL=$(MASTER_URL)
initialize: export K8S_CONFIG=$(KUBE_CONFIG)
initialize: export RUN_INTEGRATION_TESTS=true
initialize:
	@echo master URL - ${K8S_MASTER_URL}
	@echo creating executor namespace
	kubectl create ns forwarder || exit 0
	@echo launching nginx pod
	kubectl run nginx --image=nginx --port=8080 -n forwarder || exit 0

.PHONY: run
run: export K8S_MASTER_URL=$(MASTER_URL)
run: export K8S_CONFIG=$(KUBE_CONFIG)
run: export RUN_INTEGRATION_TESTS=true
run:
	@echo master URL - ${K8S_MASTER_URL}
	@echo creating executor namespace
	kubectl create ns forwarder || exit 0
	@echo launching nginx pod
	kubectl run nginx --image=nginx --port=8080 -n forwarder || exit 0
	go run example/example.go

.PHONY: test/integration
test/integration:
test/integration: export K8S_MASTER_URL=$(MASTER_URL)
test/integration: export K8S_CONFIG=$(KUBE_CONFIG)
test/integration: export RUN_INTEGRATION_TESTS=true
test/integration:
	@echo running integtation tests on $(K8S_MASTER_URL)
	@echo creating executor namespace
	kubectl create ns forwarder || exit 0
	@echo launching nginx pod
	kubectl run nginx --image=nginx --port=8080 -n forwarder || exit 0
	go test -v ./... -count=1
