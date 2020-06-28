SOURCE = $(wildcard *.go)

.PHONY: all build clean image check vendor dependencies

NAME				:= go-crond
GIT_TAG				:= $(shell git describe --dirty --tags --always)
GIT_COMMIT			:= $(shell git rev-parse --short HEAD)
LDFLAGS             := -X "main.gitTag=$(GIT_TAG)" -X "main.gitCommit=$(GIT_COMMIT)" -extldflags "-static"

PKGS				:= $(shell go list ./... | grep -v -E '/vendor/|/test')
FIRST_GOPATH		:= $(firstword $(subst :, ,$(shell go env GOPATH)))
GOLANGCI_LINT_BIN	:= $(FIRST_GOPATH)/bin/golangci-lint

GOBUILD_OSX = go build --ldflags '-w ${LDFLAGS}'
GOBUILD_DYNAMIC = go build --ldflags '\''-w ${LDFLAGS}'\''
GOBUILD_STATIC = go build --ldflags '\''-w ${LDFLAGS}'\''
.PHONY: docker docker-dev docker-run-dev all build test clean release dependencies


all: docker-dev lint build

build:
	CGO_ENABLED=0 go build -a -ldflags '$(LDFLAGS)' -o $(NAME) .

vendor:
	go mod tidy
	go mod vendor
	go mod verify

.PHONY: lint
lint: $(GOLANGCI_LINT_BIN)
	# megacheck fails to respect build flags, causing compilation failure during linting.
	# instead, use the unused, gosimple, and staticcheck linters directly
	$(GOLANGCI_LINT_BIN) run -D megacheck -E unused,gosimple,staticcheck --timeout=10m

$(GOLANGCI_LINT_BIN):
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(FIRST_GOPATH)/bin v1.23.8

ALL = \
	$(foreach arch,64 32,\
	$(foreach suffix,linux osx,\
		build/go-crond-$(arch)-$(suffix))) \
	$(foreach arch,arm arm64,\
		build/go-crond-$(arch)-linux)


docker:
	docker build . -t webdevops/go-crond

docker-dev:
	docker build -f Dockerfile.develop . -t webdevops/go-crond:develop

docker-run: docker-dev
	docker run -ti --rm -w "$$(pwd)" -v "$$(pwd):$$(pwd):ro" --name=cron webdevops/go-crond:develop bash

build-env: docker-dev
	docker run -ti --rm -w "$$(pwd)" -v "$$(pwd):$$(pwd)" --name=cron webdevops/go-crond:develop bash

all: test build

dependencies:
	go mod vendor

build: clean dependencies test $(ALL)

# cram is a python app, so 'easy_install/pip install cram' to run tests
test:
	echo test todo
	#cram tests/main.test

clean:
	rm -rf build/

# os is determined as thus: if variable of suffix exists, it's taken, if not, then
# suffix itself is taken
osx = darwin

build/go-crond-64-osx: $(SOURCE)
	@mkdir -p $(@D)
	CGO_ENABLED=1 GOOS=$(firstword $($*) $*) GOARCH=amd64 $(GOBUILD_OSX) -ldflags '$(LDFLAGS)' -o $@

build/go-crond-32-osx: $(SOURCE)
	echo "32-osx disabled"

build/go-crond-64-linux: $(SOURCE)
	@mkdir -p $(@D)
	docker run -ti --rm -w "$$(pwd)" -v "$$(pwd):$$(pwd)" --name=cron webdevops/go-crond:develop sh -c 'CGO_ENABLED=0 GOOS=$(firstword $($*) $*) GOARCH=amd64 $(GOBUILD_DYNAMIC) -o ${@}-dynamic'
	docker run -ti --rm -w "$$(pwd)" -v "$$(pwd):$$(pwd)" --name=cron webdevops/go-crond:develop sh -c 'CGO_ENABLED=0 GOOS=$(firstword $($*) $*) GOARCH=amd64 $(GOBUILD_STATIC) -o ${@}'

build/go-crond-32-linux: $(SOURCE)
	@mkdir -p $(@D)
	docker run -ti --rm -w "$$(pwd)" -v "$$(pwd):$$(pwd)" --name=cron webdevops/go-crond:develop sh -c 'CGO_ENABLED=0 GOOS=$(firstword $($*) $*) GOARCH=386 $(GOBUILD_DYNAMIC) -o ${@}-dynamic'
	docker run -ti --rm -w "$$(pwd)" -v "$$(pwd):$$(pwd)" --name=cron webdevops/go-crond:develop sh -c 'CGO_ENABLED=0 GOOS=$(firstword $($*) $*) GOARCH=386 $(GOBUILD_STATIC) -o ${@}'

build/go-crond-arm-linux: $(SOURCE)
	@mkdir -p $(@D)
	docker run -ti --rm -w "$$(pwd)" -v "$$(pwd):$$(pwd)" --name=cron webdevops/go-crond:develop sh -c 'CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=6 $(GOBUILD_DYNAMIC) -o ${@}-dynamic'
	docker run -ti --rm -w "$$(pwd)" -v "$$(pwd):$$(pwd)" --name=cron webdevops/go-crond:develop sh -c 'CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=6 $(GOBUILD_STATIC) -o ${@}'

build/go-crond-arm64-linux: $(SOURCE)
	@mkdir -p $(@D)
	docker run -ti --rm -w "$$(pwd)" -v "$$(pwd):$$(pwd)" --name=cron webdevops/go-crond:develop sh -c 'CC=aarch64-linux-gnu-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm64 $(GOBUILD_STATIC) -o ${@}'
	docker run -ti --rm -w "$$(pwd)" -v "$$(pwd):$$(pwd)" --name=cron webdevops/go-crond:develop sh -c 'CC=aarch64-linux-gnu-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm64 $(GOBUILD_DYNAMIC) -o ${@}-dynamic'


release:
	github-release release -u webdevops -r go-crond -t "$(GIT_TAG)" -n "$(GIT_TAG)" --description "$(GIT_TAG)"
	@for x in build/*; do \
		echo "Uploading $$x" && \
		github-release upload -u webdevops \
                              -r go-crond \
                              -t $(GIT_TAG) \
                              -f "$$x" \
                              -n "$$(basename $$x)"; \
	done
