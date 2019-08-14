NAME=tun2socks
BUILDDIR=$(shell pwd)/build
CMDDIR=$(shell pwd)/cmd
VERSION=$(shell git describe --tags --long || echo "unknown version")
BUILD_TAGS='fakeDNS stats'
BUILD_LDFLAGS='-s -w -X "main.version=$(VERSION)"'
DEBUG_BUILD_LDFLAGS='-s -w -X "main.version=$(VERSION)-debug"'

all: build

build:
	cd $(CMDDIR) && go build -ldflags $(BUILD_LDFLAGS) -v -tags $(BUILD_TAGS) -o $(BUILDDIR)/$(NAME)

debug:
	cd $(CMDDIR) && go build -ldflags $(DEBUG_BUILD_LDFLAGS) -v -tags $(BUILD_TAGS) -race -o $(BUILDDIR)/$(NAME)

clean:
	rm -rf $(BUILDDIR)