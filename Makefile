SOURCE = $(wildcard *.go)
TAG ?= $(shell git describe --tags)
GOBUILD = go build -ldflags '-w'
.PHONY: docker docker-dev docker-run-dev all build test clean release

ALL = \
	$(foreach arch,64 32,\
	$(foreach suffix,linux osx,\
		build/go-crond-$(arch)-$(suffix)))

docker:
	docker build . -t webdevops/go-crond

docker-dev:
	docker build -f Dockerfile.develop . -t webdevops/go-crond:develop

docker-run-dev: docker-dev
	docker run -ti --rm webdevops/go-crond:develop sh

all: test build

build: clean test $(ALL)

# cram is a python app, so 'easy_install/pip install cram' to run tests
test:
	echo test todo
	#cram tests/main.test

clean:
	rm -f $(ALL)

# os is determined as thus: if variable of suffix exists, it's taken, if not, then
# suffix itself is taken
win.exe = windows
osx = darwin
build/go-crond-64-%: $(SOURCE)
	@mkdir -p $(@D)
	CGO_ENABLED=0 GOOS=$(firstword $($*) $*) GOARCH=amd64 $(GOBUILD) -o $@

build/go-crond-32-%: $(SOURCE)
	@mkdir -p $(@D)
	CGO_ENABLED=0 GOOS=$(firstword $($*) $*) GOARCH=386 $(GOBUILD) -o $@

release: $(ALL)
	github-release release -u webdevops -r go-crond -t "$(TAG)" -n "$(TAG)" --description "$(TAG)"
	@for x in $(ALL); do \
		echo "Uploading $$x" && \
		github-release upload -u webdevops \
                              -r go-crond \
                              -t $(TAG) \
                              -f "$$x" \
                              -n "$$(basename $$x)"; \
	done
