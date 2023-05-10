BUILD_FLAGS=-trimpath -ldflags "-s -w"
GO_BUILD=CGO_ENABLED=0 go build $(BUILD_FLAGS)
MOCKERY := $(shell command -v mockery 2> /dev/null)
GOLANGCI_LINT := $(shell command -v golangci-lint 2> /dev/null)

PKG=github.com/aws/amazon-eks-connector
IMAGE_NAME?=eks-connector/eks-connector
IMAGE?=public.ecr.aws/$(IMAGE_NAME)
BIN_PATH=./bin
BUILD_PATH=./cmd
SRC_PATH=./...

.DEFAULT_GOAL := eks-connector

.PHONY: eks-connector
eks-connector: pre-compile compile

.PHONY: pre-compile
pre-compile:: lint vet imports-check-no-vendor coverage

.PHONY: lint
lint:
ifndef GOLANGCI_LINT
	$(error "golangci-lint not found (`follow https://golangci-lint.run/usage/install/#local-installation` to fix)")
endif
	GO111MODULE=on $(GOLANGCI_LINT) run

.PHONY: vet
vet::
	go vet $(SRC_PATH)

.PHONY: coverage
coverage:
	go test -coverprofile=coverage.out $(SRC_PATH)
	go tool cover -html=coverage.out

.PHONY: imports-check-no-vendor
imports-check-no-vendor:
	$(eval DIFFS := $(shell goimports -l pkg cmd))
	if [ -n "$(DIFFS)" ]; then echo "Imports or code is incorrectly formatted/ordered."; echo "Incorrectly formatted files: $(DIFFS)"; exit 1; fi

.PHONY: compile
compile::
	GOOS=windows GOARCH=amd64 $(GO_BUILD) -o $(BIN_PATH)/amd64/windows/eks-connector $(BUILD_PATH)
	GOOS=darwin GOARCH=amd64 $(GO_BUILD) -o $(BIN_PATH)/amd64/darwin/eks-connector $(BUILD_PATH)
	GOOS=darwin GOARCH=arm64 $(GO_BUILD) -o $(BIN_PATH)/arm64/darwin/eks-connector $(BUILD_PATH)
	GOOS=linux GOARCH=amd64 $(GO_BUILD) -o $(BIN_PATH)/amd64/linux/eks-connector $(BUILD_PATH)
	GOOS=linux GOARCH=arm64 $(GO_BUILD) -o $(BIN_PATH)/arm64/linux/eks-connector $(BUILD_PATH)

.PHONY: docker
docker:
	@echo 'Building image $(IMAGE)...'
	docker build -t $(IMAGE) -f configuration/Dockerfile .

.PHONY: clean
clean:: mocks-clean
	rm -rf bin/ pkg/

.PHONY: mocks-clean
mocks-clean:
	rm -f ./pkg/**/mock_*.go

.PHONY: mocks-gen
mocks-gen: mocks-clean
ifndef MOCKERY
	$(error "mockery not found (`follow https://github.com/vektra/mockery#installation` to fix)")
endif
	$(MOCKERY) \
      --case underscore \
      --all \
      --dir pkg \
      --inpackage
