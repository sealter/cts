NAME = cts

all: dep test build

dep:
	-go get -u github.com/golang/dep/cmd/dep
	-dep ensure

test:
	for d in $(shell go list ./... | grep -v vendor); do \
		go test -v $$d || exit 1; \
	done

cover:
	echo "" > coverage.txt
	for d in $(shell go list ./... | grep -v vendor); do \
		go test -race -coverprofile=profile.out -covermode=atomic $$d || exit 1; \
		[ -f profile.out ] && cat profile.out >> coverage.txt && rm profile.out; \
	done

build:
	go build -o bin/${NAME}

.PHONY: all dep test build
