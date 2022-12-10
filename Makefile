VERSION     ?=$(shell cat .version)

all: help

version: ## Prints the current version
	@echo $(VERSION)
.PHONY: version

tidy: ## Updates the go modules and vendors all dependancies 
	go mod tidy
	go mod vendor
.PHONY: tidy

app: ## Builds local binary 
	CGO_ENABLED=0 go build -trimpath \
    -ldflags="-w -s -X main.version=${VERSION} -extldflags '-static'" \
    -a -mod vendor -o app
.PHONY: app

run: ## Runs previsouly built binary
	REDIS_IP=127.0.0.1 REDIS_PORT=6379 ./app 
.PHONY: run

post: ## Posts to local service
	curl -H "Content-Type: application/json" \
	     -X POST -s -d @tests/message.json \
         http://127.0.0.1:8080/
.PHONY: post

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

image: ## Builds new image, signs it, gens SBOM, vlun report, and pushes it
	bin/image
.PHONY: image

update: ## Deploys latest image to Cloud Run
	bin/update
.PHONY: update

tag: ## Creates release tag 
	git tag -s -m "version bump to $(VERSION)" $(VERSION)
	git push origin $(VERSION)
.PHONY: tag

help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk \
		'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
.PHONY: help
