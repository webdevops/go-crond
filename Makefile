PROJECT_NAME		:= go-crond
GIT_TAG				:= $(shell git describe --dirty --tags --always)
GIT_COMMIT			:= $(shell git rev-parse --short HEAD)
LDFLAGS				:= -X "main.gitTag=$(GIT_TAG)" -X "main.gitCommit=$(GIT_COMMIT)" -extldflags "-static"

FIRST_GOPATH			:= $(firstword $(subst :, ,$(shell go env GOPATH)))
GOLANGCI_LINT_BIN		:= $(FIRST_GOPATH)/bin/golangci-lint

.PHONY: all
all: build

.PHONY: clean
clean:
	git clean -Xfd .

.PHONY: build
build:
	CGO_ENABLED=0 go build -a -ldflags '$(LDFLAGS)' -o $(PROJECT_NAME) .

.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor
	go mod verify

.PHONY: image
image: build
	docker build -t $(PROJECT_NAME):$(GIT_TAG) .

build-push-development:
	docker build -t webdevops/$(PROJECT_NAME):development . && docker push webdevops/$(PROJECT_NAME):development

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint: $(GOLANGCI_LINT_BIN)
	$(GOLANGCI_LINT_BIN) run -E exportloopref,gofmt --timeout=10m

.PHONY: dependencies
dependencies: $(GOLANGCI_LINT_BIN)

$(GOLANGCI_LINT_BIN):
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(FIRST_GOPATH)/bin v1.32.2

## -------------------------------------

GOBUILD_OSX = go build --ldflags '-w ${LDFLAGS}'
GOBUILD_DYNAMIC = go build --ldflags '\''-w ${LDFLAGS}'\''
GOBUILD_STATIC = go build --ldflags '\''-w ${LDFLAGS}'\''

.PHONY: build-local
build-local:
	CGO_ENABLED=0 go build -a -ldflags '$(LDFLAGS)' -o $(NAME) .

.PHONY: docker
docker:
	docker build --pull -fDockerfile.ubuntu -t webdevops/go-crond .

.PHONY: docker-dev
docker-dev:
	docker build -f Dockerfile.develop . -t webdevops/go-crond:develop

.PHONY: docker-run
docker-run: docker-dev
	docker run -ti --rm -w "$$(pwd)" -v "$$(pwd):$$(pwd):ro" -p 8080:8080 -e SERVER_METRICS=1 --name=cron webdevops/go-crond:develop bash

.PHONY: docker-env
build-env: docker-dev
	docker run -ti --rm -w "$$(pwd)" -v "$$(pwd):$$(pwd)" -p 8080:8080 --name=cron webdevops/go-crond:develop bash
# os is determined as thus: if variable of suffix exists, it's taken, if not, then
# suffix itself is taken
osx = darwin

BUILD_ALL = \
	$(foreach arch,64 32,\
	$(foreach suffix,linux osx,\
		build/go-crond-$(arch)-$(suffix))) \
	$(foreach arch,arm arm64,\
		build/go-crond-$(arch)-linux)

.PHONY: build
build: clean dependencies test docker-dev $(BUILD_ALL)

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
