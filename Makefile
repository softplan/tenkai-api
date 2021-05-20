IMAGE_REPO=softplan/tenkai-api

.DEFAULT_GOAL := build
.PHONY: build container-image pre-build tag-image publish

#Build the binary
build: pre-build
	@echo "Building tenkai-api"
	GOOS_VAL=$(shell go env GOOS) GOARCH_VAL=$(shell go env GOARCH) go build -v -a -installsuffix cgo -o ./build/tenkai-api cmd/tenkai/*.go

test:
	@echo "Testing tenkai-api"
	cp app.yaml ~/
	go test -v -covermode=count -coverprofile=coverage.out $(shell go list ./... | grep -v /vendor/ | grep -v /mocks/ | grep -v pkg/service/_helm/)
	go tool cover -html=coverage.out -o coverage.html

#Build the image
container-image:
	@echo "Building docker image"
	@docker build --build-arg GOOS_VAL=$(shell go env GOOS) --build-arg GOARCH_VAL=$(shell go env GOARCH) -t $(IMAGE_REPO) -f Dockerfile --no-cache .
	@echo "Docker image build successfully"

#Pre-build checks
pre-build:
	@echo "Checking system information"
	@if [ -z "$(shell go env GOOS)" ] || [ -z "$(shell go env GOARCH)" ] ; then echo 'ERROR: Could not determine the system architecture.' && exit 1 ; fi

#Tag images
tag-image: 
	@echo 'Tagging docker image'
	@export TAG=${GITHUB_REF/refs\/heads\/}
	@export TAG=${GITHUB_REF/\//-}
	@echo ${TAG}
	@docker tag $(IMAGE_REPO) $(IMAGE_REPO):$(TAG)

#Docker push image
publish:
	@echo "Pushing docker image to repository"
	@docker login -u $(DOCKER_USERNAME) -p $(DOCKER_PASSWORD)
	@docker push $(IMAGE_REPO):$(TAG)

lint:
	@echo 'GoLang source checks'
	./srccheck/update-gofmt.sh
	./srccheck/verify-gofmt.sh
	./srccheck/verify-golint.sh
	./srccheck/verify-govet.sh
