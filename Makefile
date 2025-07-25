.PHONY: build build-all

BIN_NAME = xiaozhi-mcp-pipe
LDFLAGS = -ldflags "-s -w"

# Platform definitions
PLATFORMS = darwin/amd64 darwin/arm64 linux/amd64 linux/arm64

build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BIN_NAME)

build-all:
	$(foreach platform,$(PLATFORMS), \
		$(eval GOOS = $(word 1,$(subst /, ,$(platform)))) \
		$(eval GOARCH = $(word 2,$(subst /, ,$(platform)))) \
		CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) -o $(BIN_NAME)-$(GOOS)-$(GOARCH); \
	)