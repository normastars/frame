PROJECT_NAME := "github.com/normastars/frame/example"
PKG := "$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
LDFLAGS := "-s -w -X '$(PROJECT_NAME)/version.GitCommit=`git log | grep commit | head -1 | cut -d" " -f2 | cut -c1-8`' -X '$(PROJECT_NAME)/version.BuildGoVersion=`go version | cut -d" " -f3`' -X '$(PROJECT_NAME)/version.BuildSystem=`go version | cut -d" " -f4`'"
GOBUILD := CGO_ENABLED=0 go build
GOTEST  := CGO_ENABLED=0 go test
.PHONY: build
build: ## Build the binary file
	make clean
	make manifest
	mkdir -p bin
	$(GOBUILD) -ldflags=${LDFLAGS} -o bin/server
build-min-docker:
	make clean
	make linux
	# TODO: need fix <registername> and <imagename>
	docker build -t normastars/frame/example:v2 -f build/docker/Dockerfile .
clean: ## Remove previous build
	rm -rf bin
	rm -rf manifest.txt
dep: ## Get the dependencies
	@go mod download
local: ## Built on local env project
	make build
linux: ## Build the linux version binary file
	rm -rf bin
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags=${LDFLAGS} -o bin/server
lint: ## Lint Golang files
	@golint -set_exit_status ${PKG_LIST}
manifest:
	echo `git log | grep commit | head -1 | cut -d" " -f2` > manifest.txt
run: ## Run compiled binary
	./bin/server
dev: ## Run directly via go run (CGO disabled, no compile needed)
	CGO_ENABLED=0 go run main.go
test: ## Run unittests
	@$(GOTEST) -short ${PKG_LIST}
test-coverage: ## Run tests with coverage
	@$(GOTEST) -short -coverprofile cover.out -covermode=atomic ${PKG_LIST}
	@cat cover.out >> coverage.txt
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'