NAME=tun2socks
BINDIR=$(shell pwd)/bin
VERSION=$(shell git describe --tags --long || echo "unknown version")
BUILDTAGS='fakedns session'
GOBUILD=go build -trimpath -ldflags '-s -w  -extldflags "-static" -X "github.com/xjasonlyu/tun2socks/constant.Version=$(VERSION)"'

all: build

build:
	cd component/session && packr2
	cd cmd && $(GOBUILD) -v -tags $(BUILDTAGS) -o $(BINDIR)/$(NAME)
	cd component/session && packr2 clean

clean:
	rm -rf $(BINDIR)
