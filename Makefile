.PHONY: $(shell sed -n -e '/^$$/ { n ; /^[^ .\#][^ ]*:/ { s/:.*$$// ; p ; } ; }' $(MAKEFILE_LIST))

# App config (expected from .env)
APP ?=
API_MAJOR_VERSION ?=
PORT ?=
SONAR_PROJECTKEY ?=

# override with app specific config
include .env

# Build config
GOOS ?= linux
TAG ?= v0.0.0
BUILD ?= 0
BUILD_DATE = $(shell date +%FT%T)
TARGET_ENV ?= local

ifeq ("$(API_VERSION)", "")
  PACKAGE_PATH = $(APP)
else
  PACKAGE_PATH = $(APP)/$(API_MAJOR_VERSION)
endif

ifeq ("$(TARGET_ENV)", "master")
  DOCKER_IMAGE = github.com/myrteametrics/$(APP):$(TAG)
else
  DOCKER_IMAGE = github.com/myrteametrics/$(APP):$(TAG)-$(TARGET_ENV)
endif

# Go command options
GO111MODULE ?= on
GOPROXY ?= "https://proxy.golang.org,direct"
GOSUMDB ?= "sum.golang.org"
CGO_ENABLED ?= 0
GOINSECURE ?=
GONOSUMDB ?=
GO_OPT=GOPROXY=$(GOPROXY) GOINSECURE=$(GOINSECURE) GONOSUMDB=$(GONOSUMDB) GOSUMDB=$(GOSUMDB) GO111MODULE=$(GO111MODULE) CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS)
GO_PACKAGE ?= go list ./... | \
	grep github.com/myrteametrics/$(PACKAGE_PATH)/ | \
	grep -v -e "github.com/myrteametrics/$(PACKAGE_PATH)/docs" | \
	grep -v -e "github.com/myrteametrics/$(PACKAGE_PATH)/protobuf" | \
	grep -v -e "github.com/myrteametrics/$(PACKAGE_PATH)/internal/tests"

# Go tools
export GOBIN ?= $(shell go env GOPATH)/bin
LINT = $(GOBIN)/golangci-lint
$(LINT):
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.52.2

SWAG = $(GOBIN)/swag
$(SWAG):
	go install github.com/swaggo/swag/cmd/swag@latest


help: # Show help for each of the Makefile recipes.
	@grep -E '^[a-zA-Z0-9 -]+:.*#'  Makefile | while read -r l; do printf "\033[1;32m$$(echo $$l | cut -f 1 -d':')\033[00m:$$(echo $$l | cut -f 2- -d'#')\n"; done

download: # Download all dependencies
	GO111MODULE=$(GO111MODULE) GOSUMDB=$(GOSUMDB) go mod download

lint-version: $(LINT) # Check the golangci-lint tool version
	golangci-lint --version

lint: lint-version # Lint the code
	go mod tidy
	golangci-lint run

lint-ci: lint-version # Lint the code and export a reporting for CI
	go mod tidy
	golangci-lint run --verbose --issues-exit-code 0 --out-format checkstyle > reporting/golangci-lint.out

swag-version: $(SWAG) # Check the swag tool version
	swag --version

swag: swag-version # Generate swagger documentation
	swag init --parseDependency --generalInfo --st main.go

test-integration: # Test the code
	mkdir -p reporting
	$(GO_OPT) go test -p=1 -cover -coverpkg=$$($(GO_PACKAGE) | tr '\n' ',') -coverprofile=reporting/profile.out -json $$($(GO_PACKAGE)) > reporting/tests.json || true
	go tool cover -html=reporting/profile.out -o reporting/coverage.html
	go tool cover -func=reporting/profile.out -o reporting/coverage.txt
	cat reporting/coverage.txt

test-unit: # Test the code
	mkdir -p reporting
	$(GO_OPT) go test -p=1 -short -cover -coverpkg=$$($(GO_PACKAGE) | tr '\n' ',') -coverprofile=reporting/profile.out -json $$($(GO_PACKAGE)) > reporting/tests.json  || true
	go tool cover -html=reporting/profile.out -o reporting/coverage.html
	go tool cover -func=reporting/profile.out -o reporting/coverage.txt
	cat reporting/coverage.txt

build: # Build the executable (linux by default)
	$(GO_OPT) go build -a -trimpath -ldflags "-X main.Version=$(TAG)-$(BUILD) -X main.BuildDate=$(BUILD_DATE)" -o bin/$(APP)

run: # Run the executable
	bin/$(APP)

docker-build: # Build the executable and docker image (using multi-stages build)
	docker build -t $(DOCKER_IMAGE) -f Dockerfile .

docker-build-silent: # Build the executable and docker image (using multi-stages build)
	docker build --quiet -t $(DOCKER_IMAGE) -f Dockerfile .

docker-build-local: # Build the docker image (please ensure you used "make build" before this command)
	docker build -t $(DOCKER_IMAGE) -f local.Dockerfile .

docker-build-local-silent: # Build the executable and docker image (using multi-stages build)
	docker build --quiet -t $(DOCKER_IMAGE) -f local.Dockerfile .

docker-run: # Run the docker container
	docker run -d --name $(APP) -p $(PORT):$(PORT) $(DOCKER_IMAGE)

docker-stop: # Stop the docker container
	docker stop $(APP)

docker-rm: # Remove the docker container
	docker rm -f $(APP)

docker-push: # Push the docker image to hub.docker.com
	docker push $(DOCKER_IMAGE)

docker-save: # Export the docker image to a tar file
	docker save --output $(APP)-$(TARGET_ENV).tar $(DOCKER_IMAGE)

docker-clean: # Delete the docker image from the local cache
	docker image rm $(DOCKER_IMAGE) $(docker images -f dangling=true -q) || true

docker-run-postgres-tests:
	docker run -d --rm --name myrtea-postgres-integration-tests --network host --env POSTGRES_USER=postgres --env POSTGRES_PASSWORD=postgres --env POSTGRES_DB=postgres postgres:13

docker-stop-postgres-tests:
	docker stop myrtea-postgres-integration-tests

sonar-prep:
	$(MAKE) test-integration lint-ci sonar-push

sonar-push:
	sonar-scanner \
		-Dsonar.host.url="$(SONAR_HOST_URL)" \
		-Dsonar.login="$(SONAR_LOGIN)" \
		-Dsonar.projectKey="$(SONAR_PROJECTKEY)" \
		-Dsonar.projectName="$(APP)" \
		-Dsonar.projectVersion="$(TAG)"