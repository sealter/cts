NAME = cts

GO = go
GO_GET = $(GO) get
GO_TEST = $(GO) test -v
GO_BUILD = $(GO) build -o bin/${NAME}

all: dep test build

dep:
	-$(GO_GET) ./...

test:
	$(GO_TEST) ./...

build:
	$(GO_BUILD)

.PHONY: all dep test build
