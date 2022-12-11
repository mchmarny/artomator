VERSION    ?=$(shell cat .version)
IMG_URI    ?="us-west1-docker.pkg.dev/cloudy-demos/artomator/artomator:$(shell cat .version)"
REDIS_IP   ?=127.0.0.1
REDIS_PORT ?=6379

all: help

info: ## Prints the current version
	@echo $(VERSION)
	@echo $(IMG_URI)
.PHONY: info

tidy: ## Updates the go modules and vendors all dependancies 
	go mod tidy
	go mod vendor
.PHONY: tidy

app: ## Builds local binary 
	CGO_ENABLED=0 go build -trimpath \
    -ldflags="-w -s -X main.version=${VERSION} -extldflags '-static'" \
    -a -mod vendor -o app
.PHONY: app

redis: ## Starts local redis 
	docker run --name redis-5 -dp $(REDIS_IP):$(REDIS_PORT):$(REDIS_PORT) redis
.PHONY: redis

run: ## Runs previsouly built binary
	REDIS_IP=$(REDIS_IP) REDIS_PORT=$(REDIS_PORT) ./app 
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

build: ## Builds, signs and publishes new image
	bin/build
.PHONY: build

update: ## Updates Cloud Run service with the latest image
	bin/update
.PHONY: update

deploy: ## Configures all dependancies and deploys the prebuild image 
	@echo "Deploying ${IMG_URI}"
	bin/setup
	bin/deploy $(IMG_URI)
.PHONY: deploy

clean: ## Deletes deployed resoruces 
	bin/cleanup
.PHONY: clean

tag: ## Creates release tag 
	git tag -s -m "version bump to $(VERSION)" $(VERSION)
	git push origin $(VERSION)
.PHONY: tag

help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk \
		'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
.PHONY: help
