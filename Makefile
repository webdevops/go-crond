SOURCE = $(wildcard *.go)
TAG ?= $(shell git describe --tags)
GOBUILD = go build -ldflags '-w'
.PHONY: docker docker-dev docker-run-dev all build test clean release

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

dependences:
	go get -u github.com/robfig/cron
	go get -u github.com/jessevdk/go-flags

build: clean dependences test $(ALL)

# cram is a python app, so 'easy_install/pip install cram' to run tests
test:
	echo test todo
	#cram tests/main.test

clean:
	rm -f $(ALL)

# os is determined as thus: if variable of suffix exists, it's taken, if not, then
# suffix itself is taken
osx = darwin

build/go-crond-64-osx: $(SOURCE)
	@mkdir -p $(@D)
	CGO_ENABLED=1 GOOS=$(firstword $($*) $*) GOARCH=amd64 $(GOBUILD) -o $@

build/go-crond-32-osx: $(SOURCE)
	@mkdir -p $(@D)
	CGO_ENABLED=1 GOOS=$(firstword $($*) $*) GOARCH=386 $(GOBUILD) -o $@

build/go-crond-64-linux: $(SOURCE)
	@mkdir -p $(@D)
	docker run -ti --rm -w "$$(pwd)" -v "$$(pwd):$$(pwd)" --name=cron webdevops/go-crond:develop sh -c 'CGO_ENABLED=1 GOOS=$(firstword $($*) $*) GOARCH=amd64 $(GOBUILD) -o $@'

build/go-crond-32-linux: $(SOURCE)
	@mkdir -p $(@D)
	docker run -ti --rm -w "$$(pwd)" -v "$$(pwd):$$(pwd)" --name=cron webdevops/go-crond:develop sh -c 'CGO_ENABLED=1 GOOS=$(firstword $($*) $*) GOARCH=386 $(GOBUILD) -o $@'

build/go-crond-arm-linux: $(SOURCE)
	@mkdir -p $(@D)
	docker run -ti --rm -w "$$(pwd)" -v "$$(pwd):$$(pwd)" --name=cron webdevops/go-crond:develop sh -c 'CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=6 $(GOBUILD) -o $@'

build/go-crond-arm64-linux: $(SOURCE)
	@mkdir -p $(@D)
	docker run -ti --rm -w "$$(pwd)" -v "$$(pwd):$$(pwd)" --name=cron webdevops/go-crond:develop sh -c 'CC=aarch64-linux-gnu-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm64 $(GOBUILD) -o $@'

release: build
	github-release release -u webdevops -r go-crond -t "$(TAG)" -n "$(TAG)" --description "$(TAG)"
	@for x in $(ALL); do \
		echo "Uploading $$x" && \
		github-release upload -u webdevops \
                              -r go-crond \
                              -t $(TAG) \
                              -f "$$x" \
                              -n "$$(basename $$x)"; \
	done
