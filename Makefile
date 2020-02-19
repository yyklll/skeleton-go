NAME:=skeleton
DOCKER_REGISTRY:=docker.io
DOCKER_REPOSITORY:=yyklll
DOCKER_IMAGE_NAME:=$(DOCKER_REPOSITORY)/$(NAME)
GIT_COMMIT:=$(shell git describe --long --always --abbrev=14)
FLAGS_DEBUG:="-X main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.githash=$(GIT_COMMIT)"
FLAGS_RELEASE:="-s -w -X main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.githash=$(GIT_COMMIT)" # omit debug information
VERSION:=$(shell grep 'VERSION' pkg/version/version.go | awk '{ print $$4 }' | tr -d '"')
REVISION:=$(shell grep 'REVERSION' pkg/version/version.go | awk '{ print $$4 }' | tr -d '"')

run:
	GO111MODULE=on go run -ldflags $(FLAGS_RELEASE) cmd/server/* --level=debug

test:
	GO111MODULE=on go test -v -race ./...

build:
	GO111MODULE=on GIT_COMMIT=$$(git rev-list -1 HEAD) && GO111MODULE=on CGO_ENABLED=0 go build  -ldflags $(FLAGS_DEBUG) -a -o ./bin/server ./cmd/server/*

build-charts:
	helm lint charts/*
	helm package charts/*

build-container:
	docker build -t $(DOCKER_IMAGE_NAME):$(VERSION) .
	docker tag -t $(DOCKER_IMAGE_NAME):$(VERSION) $(DOCKER_IMAGE_NAME):latest

test-container:
	@docker rm -f $(DOCKER_IMAGE_NAME) || true
	@docker run -d -p6666:6666 --name=$(DOCKER_IMAGE_NAME) $(DOCKER_IMAGE_NAME):$(VERSION)
	@TOKEN=$$(curl -sd 'test' localhost:6666/token | jq -r .token) && \
	curl -sH "Authorization: Bearer $${TOKEN}" localhost:6666/token/validate | grep test

push-container:
	docker tag $(DOCKER_IMAGE_NAME):$(VERSION) $(DOCKER_REGISTRY)/$(DOCKER_REPOSITORY)/$(DOCKER_IMAGE_NAME):$(VERSION)
	docker tag $(DOCKER_IMAGE_NAME):$(VERSION) $(DOCKER_REGISTRY)/$(DOCKER_REPOSITORY)/$(DOCKER_IMAGE_NAME):latest
	docker push $(DOCKER_REGISTRY)/$(DOCKER_REPOSITORY)/$(DOCKER_IMAGE_NAME):$(VERSION)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_REPOSITORY)/$(DOCKER_IMAGE_NAME):latest

set-version:
    echo "Specify the next version as 'make set-version NEXT=[the next version]'"
	current="$(VERSION)" && \
	sed -i '' "s/$$current/$(NEXT)/g" pkg/version/version.go && \
	sed -i '' "s/tag: $$current/tag: $(NEXT)/g" charts/$(NAME)/values.yaml && \
	sed -i '' "s/appVersion: $$current/appVersion: $(NEXT)/g" charts/$(NAME)/Chart.yaml && \
	sed -i '' "s/version: $$current/version: $(NEXT)/g" charts/$(NAME)/Chart.yaml && \
	sed -i '' "s/$(NAME):$$current/$(NAME):$(NEXT)/g" kustomize/deployment.yaml && \
	echo "Version $(NEXT) set in code, deployment, chart and kustomize"

set-revision:
    echo "Specify the next revision as 'make set-revision NEXT=[the next revision]'"
	current="$(REVISION)" && \
	sed -i '' "s/$$current/$(NEXT)/g" pkg/version/version.go && \
	echo "Revision $(NEXT) set in code"

release:
	git tag $(VERSION)
	git push origin $(VERSION)

swagger:
	GO111MODULE=on go get github.com/swaggo/swag/cmd/swag
	cd pkg/api && $$(go env GOPATH)/bin/swag init -g server.go
