APP = myrtea-engine-api
GOOS ?= linux
TAG ?= v0.0.0
BUILD ?= 0
BUILD_DATE = $(shell date +%FT%T)
TARGET_ENV ?= local
PORT ?= 9000
TARGET_PORT ?= 9000

ifeq ("$(TARGET_ENV)", "master")
  DOCKER_IMAGE = github.com/myrteametrics/$(APP):$(TAG)
else
  DOCKER_IMAGE = github.com/myrteametrics/$(APP):$(TAG)-$(TARGET_ENV)
endif

GO111MODULE ?= on
GOSUMDB ?= off

GO_PACKAGE ?= go list ./... | \
	grep github.com/myrteametrics/myrtea-engine-api/v4/ | \
	grep -v -e "github.com/myrteametrics/myrtea-engine-api/v4/docs" | \
	grep -v -e "github.com/myrteametrics/myrtea-engine-api/v4/internals/tests"

export GOBIN ?= $(shell go env GOPATH)/bin

SWAG = $(GOBIN)/swag
LINT =$(GOBIN)/lint

.DEFAULT_GOAL := help
.PHONY: help
help:
	@echo "===  Myrtea build & deployment helper ==="
	@echo "* Don't forget to check all overridable variables"
	@echo ""
	@echo "The following commands are available :"
	@grep -E '^\.PHONY: [a-zA-Z_-]+.*?## .*$$' $(MAKEFILE_LIST) | cut -c9- | awk 'BEGIN {FS = " ## "}; {printf "\033[36m%-40s\033[0m %s\n", $$1, $$2}'


.PHONY: download ## Download all dependencies
download:
	GO111MODULE=$(GO111MODULE) GOSUMDB=$(GOSUMDB) go mod download

.PHONY: test-integration-lw ## Test the code
test-integration-lw:
	GO111MODULE=$(GO111MODULE) GOSUMDB=$(GOSUMDB) CGO_ENABLED=0 go test -p=1 $$($(GO_PACKAGE))

.PHONY: test-integration-lw-package ## Test the code
test-integration-lw-package:
	GO111MODULE=$(GO111MODULE) GOSUMDB=$(GOSUMDB) CGO_ENABLED=0 go test -p=1 github.com/myrteametrics/myrtea-engine-api/v4/internals/$(GO_PACKAGE)

.PHONY: test-integration ## Test the code
test-integration:
	mkdir -p coverage
	GO111MODULE=$(GO111MODULE) GOSUMDB=$(GOSUMDB) CGO_ENABLED=0 go test -p=1 -cover -coverpkg=$$($(GO_PACKAGE) | tr '\n' ',') -coverprofile=coverage/profile.out $$($(GO_PACKAGE))
	go tool cover -html=coverage/profile.out -o coverage/coverage.html
	go tool cover -func=coverage/profile.out -o coverage/coverage.txt
	cat coverage/coverage.txt

.PHONY: test-integration-package ## Test the code
test-integration-package:
	mkdir -p coverage
	GO111MODULE=$(GO111MODULE) GOSUMDB=$(GOSUMDB) CGO_ENABLED=0 go test -p=1 -cover -coverpkg=github.com/myrteametrics/myrtea-engine-api/v4/internals/$(GO_PACKAGE) -coverprofile=coverage/profile.out github.com/myrteametrics/myrtea-engine-api/v4/internals/$(GO_PACKAGE)
	go tool cover -html=coverage/profile.out -o coverage/coverage.html
	go tool cover -func=coverage/profile.out -o coverage/coverage.txt
	cat coverage/coverage.txt

.PHONY: test-unit ## Test the code
test-unit:
	mkdir -p coverage
	GO111MODULE=$(GO111MODULE) GOSUMDB=$(GOSUMDB) CGO_ENABLED=0 go test -p=1 -short -cover -coverpkg=$$($(GO_PACKAGE) | tr '\n' ',') -coverprofile=coverage/profile.out $$($(GO_PACKAGE))
	go tool cover -html=coverage/profile.out -o coverage/coverage.html
	go tool cover -func=coverage/profile.out -o coverage/coverage.txt
	cat coverage/coverage.txt

# .PHONY: test-race
# test-race:
# 	GO111MODULE=$(GO111MODULE) GOSUMDB=$(GOSUMDB) go test -short -race $$(go list ./... | grep -v /vendor/)

# .PHONY: test-memory
# test-memory:
# 	GO111MODULE=$(GO111MODULE) GOSUMDB=$(GOSUMDB) go test -msan -short $$(go list ./... | grep -v /vendor/)


$(GOLINT):
	go get golang.org/x/lint/golint

.PHONY: lint ## Lint the code
lint: $(GOLINT)
	golint -set_exit_status=true $$(go list ./... | grep github.com/myrteametrics/myrtea-engine-api/v4)

$(GOLANGCILINT):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.30.0

lint2: $(GOLANGCILINT)
	golangci-lint run

$(SWAG):
	go install github.com/swaggo/swag/cmd/swag@v1.5.1

.PHONY: swag ## Generate swagger documentation
swag: $(SWAG)
	swag --version
	swag init --generalInfo main.go

.PHONY: build ## Build the executable (linux by default)
build:
	GO111MODULE=$(GO111MODULE) GOSUMDB=$(GOSUMDB) CGO_ENABLED=0 GOOS=$(GOOS) go build -a -trimpath -ldflags "-X main.Version=$(TAG)-$(BUILD) -X main.BuildDate=$(BUILD_DATE)" -o bin/$(APP)

.PHONY: run ## Run the executable
run:
	bin/$(APP)

.PHONY: docker-build ## Build the executable and docker image (using multi-stages build)
docker-build:
	docker build -t $(DOCKER_IMAGE) -f Dockerfile .

.PHONY: docker-build-silent ## Build the executable and docker image (using multi-stages build)
docker-build-silent:
	docker build --quiet -t $(DOCKER_IMAGE) -f Dockerfile .

.PHONY: docker-build-local ## Build the docker image (please ensure you used "make build" before this command)
docker-build-local:
	docker build -t $(DOCKER_IMAGE) -f local.Dockerfile .

.PHONY: docker-build-local-silent ## Build the executable and docker image (using multi-stages build)
docker-build-local-silent:
	docker build --quiet -t $(DOCKER_IMAGE) -f local.Dockerfile .

.PHONY: docker-run ## Run the docker container
docker-run:
	docker run -d --name $(APP) -p $(PORT):$(TARGET_PORT) $(DOCKER_IMAGE)

.PHONY: docker-stop ## Stop the docker container
docker-stop:
	docker stop $(APP)

.PHONY: docker-rm ## Remove the docker container
docker-rm:
	docker rm -f $(APP)

.PHONY: docker-push ## Push the docker image to hub.docker.com
docker-push:
	docker push $(DOCKER_IMAGE)

.PHONY: docker-save ## Export the docker image to a tar file
docker-save:
	docker save --output $(APP)-$(TARGET_ENV).tar $(DOCKER_IMAGE)

.PHONY: docker-clean ## Delete the docker image from the local cache
docker-clean:
	docker image rm $(DOCKER_IMAGE) $(docker images -f dangling=true -q) || true