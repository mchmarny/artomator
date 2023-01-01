VERSION  =$(shell cat .version)
COMMIT   =$(shell git rev-parse HEAD)
IMG_URI  ="us-west1-docker.pkg.dev/cloudy-demos/artomator/artomator:$(shell cat .version)"
TEST_RIP =127.0.0.1
TEST_RPT =6379
TEST_PRJ =cloudy-demos
TEST_KEY =gcpkms://projects/cloudy-demos/locations/global/keyRings/artomator-signer-ring/cryptoKeys/artomator-signer
TEST_BCT =artomator-cloudy-demos
TEST_ADD =http://127.0.0.1:8080

export PATH := bin:$(PATH)

all: help

info: ## Prints the current version and iamge
	@echo $(VERSION)
	@echo $(IMG_URI)
	@echo $(COMMIT)
.PHONY: info

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

bin: ## Builds local binary 
	CGO_ENABLED=0 go build -trimpath \
    -ldflags="-w -s -X main.version=${VERSION} -extldflags '-static'" \
    -a -mod vendor -o server cmd/server/main.go
.PHONY: bin

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
	./server 
.PHONY: server

event: ## Submits events test to local service
	tools/run-event $(TEST_ADD)
.PHONY: event-test

sbom: image ## Submits SBOM process request test to local service
	tools/run-sbom $(TEST_ADD)
.PHONY: sbom-test 

docker-run: ## Runs bash on latest artomator image
	docker container run --rm -it --entrypoint /bin/bash $(IMG_URI)
.PHONY: docker-run

docker-clean: ## Removes orpaned docker volumes
	@echo "stopping all containers..."
	docker stop $(shell docker ps -aq)
	@echo "removing all containers..." 
	docker rm $(shell docker ps -aq)
	@echo "prunning system..."
	docker system prune -a --volumes
	@echo "done"
.PHONY: docker-clean

image: ## Makes test image 
	tools/test-image
.PHONY: image

build: ## Builds, signs and publishes new image locally (dirty)
	tools/build
.PHONY: build

release: test lint tag ## Runs test, lint, and tag before release
	@echo "Releasing: $(VERSION)"
	tools/gh-wait
	tools/tf-apply
.PHONY: release

attest: ## Checks attestation for the current image using cosign
	cosign verify-attestation --key $(TEST_KEY) --type spdx \
		$(shell cat tests/test-digest.txt) | \
		jq -r .payload | base64 -D | jq .predicateType
.PHONY: attest

policy: ## Creates k8s admission policies based on the current config
	tools/make-policy
.PHONY: policy

policy-controller: ## Downloads and configures latest policy controller
	wget "https://github.com/sigstore/policy-controller/releases/download/v0.5.2/tester-darwin-arm64"
	sudo mv tester-darwin-arm64 /usr/local/bin/policy-tester
	sudo chmod +x /usr/local/bin/policy-tester
.PHONY: policy-controller

policy-test: ## Tests generated policy using policy tester
	../policy-controller/policy-tester --policy tests/policy/image-policy.yaml --image $(shell cat tests/test-digest.txt)
.PHONY: policy-test

infra: ## Applies Terraform 
	terraform -chdir=./deployment apply -auto-approve
.PHONY: infra

infra-fmt: ## Formats Terraform 
	terraform -chdir=./deployment fmt
.PHONY: nice

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
