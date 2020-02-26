NAME=tun2socks
BINDIR=$(shell pwd)/bin
VERSION=$(shell git describe --tags --long || echo "unknown version")
BUILDTAGS='fakedns session'
GOBUILD=go build -ldflags '-s -w -X "github.com/xjasonlyu/tun2socks/constant.Version=$(VERSION)"'

all: build

build:
	cd cmd && $(GOBUILD) -v -tags $(BUILDTAGS) -o $(BINDIR)/$(NAME)

debug:
	cd cmd && $(GOBUILD) -v -tags $(BUILDTAGS) -race -o $(BINDIR)/$(NAME)

clean:
	rm -rf $(BINDIR)
