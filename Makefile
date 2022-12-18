VERSION  =$(shell cat .version)
IMG_URI  ="us-west1-docker.pkg.dev/cloudy-demos/artomator/artomator:$(shell cat .version)"
TEST_RIP =127.0.0.1
TEST_RPT =6379
TEST_PRJ =cloudy-demos
TEST_KEY =gcpkms://projects/cloudy-demos/locations/us-west1/keyRings/artomator/cryptoKeys/artomator-signer/cryptoKeyVersions/1
TEST_BCT =cloudy-demos-artomator

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
    -a -mod vendor -o app cmd/server/main.go
.PHONY: app

redis: ## Starts local redis 
	docker run --name redis-5 -dp $(TEST_RIP):$(TEST_RPT):$(TEST_RPT) redis
.PHONY: redis

redis-stop: ## Stops and removes running local redis 
	docker rm /redis-5
.PHONY: redis-stop

server: ## Runs previsouly built server binary
	PROJECT_ID=$(TEST_PRJ) SIGN_KEY=$(TEST_KEY) \
	REDIS_IP=$(TEST_RIP) REDIS_PORT=$(TEST_RPT) \
	GCS_BUCKET=$(TEST_BCT) \
	./app 
.PHONY: server

event-test: ## Submits events test to local service
	curl -i -X POST -H "Content-Type: application/json" \
	     -s -d @tests/message.json \
         "http://127.0.0.1:8080/event"
.PHONY: post

process-test: ## Submits process test to local service
	curl -i -X POST -H "Content-Type: application/json" \
         "http://127.0.0.1:8080/process?digest=$(shell cat tests/test-digest.txt)"
.PHONY: patch

verify-test: ## Submits verify test to local service
	curl -i -X POST -H "Content-Type: application/json" \
         "http://127.0.0.1:8080/verify?format=spdx&digest=$(shell cat tests/test-digest.txt)"
.PHONY: get

scan-test: ## Submits scan test to local service
	curl -i -X POST -H "Content-Type: application/json" \
         "http://127.0.0.1:8080/scan?severity=low&scope=squashed&digest=$(shell cat tests/test-digest.txt)"
.PHONY: get

cmd: ## Runs bash on latest artomator image
	docker container run --rm -it --entrypoint /bin/bash $(IMG_URI)
.PHONY: cmd

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

image: ## Makes test image 
	tests/build-test-image
.PHONY: image

build: ## Builds, signs and publishes new image
	tools/build
.PHONY: build

update: ## Updates Cloud Run service with the latest image
	tools/update
.PHONY: update

deploy: ## Configures all dependancies and deploys the prebuild image 
	@echo "Deploying ${IMG_URI}"
	tools/setup
	tools/deploy $(IMG_URI)
.PHONY: deploy

clean: ## Deletes deployed resoruces 
	tools/cleanup
.PHONY: clean

docker-clean: ## Removes orpaned docker volumes
	@echo "stopping all containers..."
	docker stop $(shell docker ps -aq)
	@echo "removing all containers..." 
	docker rm $(shell docker ps -aq)
	@echo "prunning system..."
	docker system prune -a --volumes
	@echo "done"
.PHONY: docker-clean

tag: ## Creates release tag 
	git tag -s -m "version bump to $(VERSION)" $(VERSION)
	git push origin $(VERSION)
.PHONY: tag

tagless: ## Delete the current release tag 
	git tag -d $(VERSION)
	git push --delete origin $(VERSION)
.PHONY: tagless

help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk \
		'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
.PHONY: help
