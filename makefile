NAME = cts

GO = go
GO_GET = $(GO) get
GO_TEST = $(GO) test -v
GO_BUILD = $(GO) build -o bin/${NAME}
GO_INSTALL = $(GO) install

all: dep test build install

dep:
	-$(GO_GET) github.com/smartystreets/goconvey/convey
	-$(GO_GET) github.com/urfave/cli

test:
	$(GO_TEST)

build:
	$(GO_BUILD)

install:
	$(GO_INSTALL)

.PHONY: all dep test build install
