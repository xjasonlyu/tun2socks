GOCMD=go
XGOCMD=xgo
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
VERSION=$(shell git describe --tags)
DEBUG_LDFLAGS=''
RELEASE_LDFLAGS='-s -w -X main.version=$(VERSION)'
BUILD_TAGS?=dnscache socks dnsfallback fakedns d stats
DEBUG_BUILD_TAGS=$(BUILD_TAGS) debug
BUILDDIR=$(shell pwd)/build
CMDDIR=$(shell pwd)/cmd/tun2socks
PROGRAM=tun2socks

BUILD_CMD="cd $(CMDDIR) && $(GOBUILD) -ldflags $(RELEASE_LDFLAGS) -o $(BUILDDIR)/$(PROGRAM) -v -tags '$(BUILD_TAGS)'"
DBUILD_CMD="cd $(CMDDIR) && $(GOBUILD) -race -ldflags $(DEBUG_LDFLAGS) -o $(BUILDDIR)/$(PROGRAM) -v -tags '$(DEBUG_BUILD_TAGS)'"
XBUILD_CMD="cd $(BUILDDIR) && $(XGOCMD) -ldflags $(RELEASE_LDFLAGS) -tags '$(BUILD_TAGS)' --targets=*/* $(CMDDIR)"
RELEASE_CMD="cd $(BUILDDIR) && $(XGOCMD) -ldflags $(RELEASE_LDFLAGS) -tags '$(BUILD_TAGS)' --targets=linux/amd64,linux/arm64,linux/386,linux/mips,linux/mipsle,linux/mips64,linux/mips64le,windows/*,darwin/* $(CMDDIR)"
WINDOWS_CMD="cd $(BUILDDIR) && $(XGOCMD) -ldflags $(RELEASE_LDFLAGS) -tags '$(BUILD_TAGS)' --targets=windows/amd64 $(CMDDIR)"

all: build

build:
	mkdir -p $(BUILDDIR)
	eval $(BUILD_CMD)

dbuild:
	mkdir -p $(BUILDDIR)
	eval $(DBUILD_CMD)

windows:
	mkdir -p $(BUILDDIR)
	eval $(WINDOWS_CMD)

xbuild:
	mkdir -p $(BUILDDIR)
	eval $(XBUILD_CMD)

release:
	mkdir -p $(BUILDDIR)
	eval $(RELEASE_CMD)

travisbuild: xbuild

clean:
	rm -rf $(BUILDDIR)

cleancache:
	# go build cache may need to cleanup if changing C source code
	$(GOCLEAN) -cache
	rm -rf $(BUILDDIR)
