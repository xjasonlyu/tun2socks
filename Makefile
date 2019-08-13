NAME=tun2socks
BUILDDIR=$(shell pwd)/build
CMDDIR=$(shell pwd)/cmd/tun2socks
VERSION=$(shell git describe --tags --long || echo "unknown version")
BUILD_TAGS='fakedns socks stats'
BUILD_LDFLAGS='-s -w -X "main.version=$(VERSION)"'
GOBUILD=go build -ldflags $(BUILD_LDFLAGS) -v -tags $(BUILD_TAGS)

all: build

build:
	cd $(CMDDIR) && $(GOBUILD) -o $(BUILDDIR)/$(NAME)

debug:
	cd $(CMDDIR) && $(GOBUILD) -race -o $(BUILDDIR)/$(NAME)

clean:
	rm -rf $(BUILDDIR)