IMAGE_REPO=softplan/tenkai-api
TAG=$(TRAVIS_BRANCH)

.DEFAULT_GOAL := build
.PHONY: release git-tag check-git-status build container-image pre-build tag-image publish

#Docker Tasks
#Make a release
release: check-git-status container-image tag-image publish git-tag 
	@echo "Successfully released version $(TAG)"

#Create a git tag
git-tag:
	@echo "Creating a git tag"
	@git add .release
	@git commit -m "Release $(TAG)" ;
	@git tag $(TAG) ;
	@git push --tags origin develop;
	@echo 'Git tag pushed successfully' ;

#Check git status
check-git-status:
	@echo "Checking git status"
	@if [ -n "$(shell git tag | grep $(TAG))" ] ; then echo 'ERROR: Tag already exists' && exit 1 ; fi
	@if [ -z "$(shell git remote -v)" ] ; then echo 'ERROR: No remote to push tags to' && exit 1 ; fi
	@if [ -z "$(shell git config user.email)" ] ; then echo 'ERROR: Unable to detect git credentials' && exit 1 ; fi

#Build the binary
build: pre-build
	@echo "Building tenkai-api"
	GOOS_VAL=$(shell go env GOOS) GOARCH_VAL=$(shell go env GOARCH) go build -a -installsuffix cgo -o ./build/tenkai-api cmd/tenkai/*.go

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
	@echo 'Tagging image'
	@docker tag $(IMAGE_REPO) $(IMAGE_REPO):$(TAG)

#Docker push image
publish:
	@echo "Pushing docker image to repository"
	@docker login -u $(DOCKER_USERNAME) -p $(DOCKER_PASSWORD)
	@docker push $(IMAGE_REPO):$(TAG)
