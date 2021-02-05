BINDIR = "bin"
NAME = "tun2socks"

TAGS = ""
BUILD_FLAGS = "-v"

VERSION = $(shell git describe --tags || echo "unknown version")
BUILD_TIME = $(shell date -u '+%FT%TZ')

LDFLAGS += -w -s -buildid=
LDFLAGS += -X "github.com/xjasonlyu/tun2socks/constant.Version=$(VERSION)"
LDFLAGS += -X "github.com/xjasonlyu/tun2socks/constant.BuildTime=$(BUILD_TIME)"  # RFC3339
GO_BUILD = GO111MODULE=on CGO_ENABLED=0 go build $(BUILD_FLAGS) -ldflags '$(LDFLAGS)' -tags '$(TAGS)' -trimpath

PLATFORM_LIST = \
	darwin-amd64 \
	freebsd-amd64 \
	freebsd-arm64 \
	linux-amd64 \
	linux-arm64 \
	openbsd-amd64 \
	openbsd-arm64 \

WINDOWS_ARCH_LIST = \
	windows-amd64 \

all: linux-amd64 darwin-amd64 windows-amd64

.PHONY: docker
docker:
	$(GO_BUILD) -o $(BINDIR)/$(NAME)-$@

darwin-amd64:
	GOARCH=amd64 GOOS=darwin $(GO_BUILD) -o $(BINDIR)/$(NAME)-$@

freebsd-amd64:
	GOARCH=amd64 GOOS=freebsd $(GO_BUILD) -o $(BINDIR)/$(NAME)-$@

freebsd-arm64:
	GOARCH=arm64 GOOS=freebsd $(GO_BUILD) -o $(BINDIR)/$(NAME)-$@

linux-amd64:
	GOARCH=amd64 GOOS=linux $(GO_BUILD) -o $(BINDIR)/$(NAME)-$@

linux-arm64:
	GOARCH=arm64 GOOS=linux $(GO_BUILD) -o $(BINDIR)/$(NAME)-$@

openbsd-amd64:
	GOARCH=amd64 GOOS=openbsd $(GO_BUILD) -o $(BINDIR)/$(NAME)-$@

openbsd-arm64:
	GOARCH=arm64 GOOS=openbsd $(GO_BUILD) -o $(BINDIR)/$(NAME)-$@

windows-amd64:
	GOARCH=amd64 GOOS=windows $(GO_BUILD) -o $(BINDIR)/$(NAME)-$@.exe

gz_releases=$(addsuffix .gz, $(PLATFORM_LIST))
zip_releases=$(addsuffix .zip, $(WINDOWS_ARCH_LIST))

$(gz_releases): %.gz : %
	chmod +x $(BINDIR)/$(NAME)-$(basename $@)
	gzip -f -S .gz $(BINDIR)/$(NAME)-$(basename $@)

$(zip_releases): %.zip : %
	zip -m -j $(BINDIR)/$(NAME)-$(basename $@).zip $(BINDIR)/$(NAME)-$(basename $@).exe

all-arch: $(PLATFORM_LIST) $(WINDOWS_ARCH_LIST)

releases: $(gz_releases) $(zip_releases)

clean:
	rm $(BINDIR)/*
