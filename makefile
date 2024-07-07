GOCMD=go
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
BINARY_NAME=go-rest-template
VERSION?=1.0.0
DOCKER_REGISTRY?= #if set it should finished by /

# determine if running on Linux
ifneq ($(OS),Windows_NT)
	GREEN  := $(shell tput -Txterm setaf 2)
	YELLOW := $(shell tput -Txterm setaf 3)
	WHITE  := $(shell tput -Txterm setaf 7)
	CYAN   := $(shell tput -Txterm setaf 6)
	RESET  := $(shell tput -Txterm sgr0)
	GOHOME := $(HOME)/go/bin/
endif

.PHONY: all test build clean coverage lint lint-go vet-go docker-build docker-release help

all: help

## Run:
run: ## Run the project
	$(GOCMD) run .

## Build:
build: ## Build your project and put the output binary in out/bin/
	$(GOCMD) build -o out/bin/$(BINARY_NAME) .

clean: ## Remove build related file
ifeq ($(OS),Windows_NT)
	del /q /s .\out
	del /q /s .\junit-report.xml
	del /q /s .\junit-raw.txt
	del /q /s .\checkstyle-report.xml
	del /q /s .\coverage.xml
	del /q /s .\profile.json
	del /q /s .\profile.cov
	rmdir /q /s .\out
else
	rm -fr ./bin
	rm -fr ./out
	rm -f ./junit-raw.txt ./junit-report.xml checkstyle-report.xml ./coverage.xml ./profile.json ./profile.cov
endif

## Test:
test: ## Run the tests of the project
	$(GOTEST) -v -race ./...

test-junit: ## Run the tests of the project and export a junit report
	go install github.com/jstemmer/go-junit-report@latest
	$(GOTEST) -v -race 2>&1 ./... > junit-raw.txt
	$(GOHOME)go-junit-report -set-exit-code < junit-raw.txt > junit-report.xml

coverage: ## Run the tests of the project and display coverage
	$(GOTEST) -cover -covermode=count -coverprofile=profile.cov ./...
	$(GOCMD) tool cover -func profile.cov

cobertura: ## Run the tests of the project and export a cobertura coverage xml
	go install github.com/axw/gocov/gocov@latest
	go install github.com/AlekSi/gocov-xml@latest
	$(GOTEST) -cover -covermode=count -coverprofile=profile.cov ./...
	$(GOCMD) tool cover -func profile.cov
	$(GOHOME)gocov convert profile.cov > profile.json
	$(GOHOME)gocov-xml < profile.json > coverage.xml

## Lint:
lint: vet-go lint-go ## Run all available linters

lint-go: ## Use staticcheck on your project
	go install honnef.co/go/tools/cmd/staticcheck@latest
	$(GOHOME)staticcheck ./...

vet-go: ## Use go vet on your project
	$(GOVET)

## Docker:
docker-build: ## Use the dockerfile to build the container
	docker build --rm --tag $(BINARY_NAME) .

docker-run: ## Use docker compose to run the project
	docker compose up --detached

docker-stop: ## Use docker compose to stop the running project
	docker compose down

docker-release: ## Release the container with tag latest and version
	docker tag $(BINARY_NAME) $(DOCKER_REGISTRY)$(BINARY_NAME):latest
	docker tag $(BINARY_NAME) $(DOCKER_REGISTRY)$(BINARY_NAME):$(VERSION)
	docker push $(DOCKER_REGISTRY)$(BINARY_NAME):latest
	docker push $(DOCKER_REGISTRY)$(BINARY_NAME):$(VERSION)

## Help:
help: ## Show this help.
ifeq ($(OS),Windows_NT)
	@echo Usage:
	@echo make target
else
	@echo ''
	@echo 'Usage:'
	@echo '${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)
endif