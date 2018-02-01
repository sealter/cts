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

release:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/${NAME}
	sudo docker build -t modood/cts .
	sudo docker push modood/cts

.PHONY: all dep test cover build release
