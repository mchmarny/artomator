VERSION     ?=$(shell cat .version)
COMMIT      ?=$(shell git rev-parse --short HEAD)

all: help

version: ## Prints the current version
	@echo $(VERSION)
.PHONY: version

tidy: ## Updates the go modules and vendors all dependancies 
	go mod tidy
	go mod vendor
.PHONY: tidy

upgrade: ## Upgrades all dependancies 
	go get -d -u ./...
	go mod tidy
	go mod vendor
.PHONY: upgrade

test: tidy ## Runs unit tests
	go test -short -count=1 -race -covermode=atomic -coverprofile=cover.out ./...
.PHONY: test

lint: ## Lints the entire project 
	golangci-lint -c .golangci.yaml run
.PHONY: lint

run: ## Runs uncompiled app 
	go run main.go
.PHONY: run

post: ## Post event to local endpoint 
	curl -s -H "Content-Type: application/json" \
    	 -d @events/invalid.json \
		 -X POST http://localhost:8080 | jq "."
	curl -s -H "Content-Type: application/json" \
    	 -d @events/tag.json \
		 -X POST http://localhost:8080 | jq "."
	curl -s -H "Content-Type: application/json" \
    	 -d @events/valid.json \
		 -X POST http://localhost:8080 | jq "."
.PHONY: post

image: ## Build local image using Docker
	bin/image
.PHONY: image

tag: ## Creates release tag 
	git tag -s -m "version bump to $(VERSION)" $(VERSION)
	git push origin $(VERSION)
.PHONY: tag

setup:
	bin/setup
.PHONY: setup 

help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk \
		'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
.PHONY: help
