BINARY = tun2socks
OUTPUT = bin

BUILD_FLAGS   = -v
BUILD_TAGS    =
BUILD_COMMIT  = $(shell git describe --dirty --always)
BUILD_VERSION = $(shell git describe --abbrev=0 --tags HEAD)

CGO_ENABLED = 0
GO111MODULE = on

LDFLAGS += -w -s -buildid=
LDFLAGS += -X "github.com/xjasonlyu/tun2socks/constant.Version=$(BUILD_VERSION)"
LDFLAGS += -X "github.com/xjasonlyu/tun2socks/constant.GitCommit=$(BUILD_COMMIT)"

GO_BUILD = GO111MODULE=$(GO111MODULE) CGO_ENABLED=$(CGO_ENABLED) \
	go build $(BUILD_FLAGS) -ldflags '$(LDFLAGS)' -tags '$(BUILD_TAGS)' -trimpath

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

tun2socks:
	$(GO_BUILD) -o $(OUTPUT)/$(BINARY)

darwin-amd64:
	GOARCH=amd64 GOOS=darwin $(GO_BUILD) -o $(OUTPUT)/$(BINARY)-$@

freebsd-amd64:
	GOARCH=amd64 GOOS=freebsd $(GO_BUILD) -o $(OUTPUT)/$(BINARY)-$@

freebsd-arm64:
	GOARCH=arm64 GOOS=freebsd $(GO_BUILD) -o $(OUTPUT)/$(BINARY)-$@

linux-amd64:
	GOARCH=amd64 GOOS=linux $(GO_BUILD) -o $(OUTPUT)/$(BINARY)-$@

linux-arm64:
	GOARCH=arm64 GOOS=linux $(GO_BUILD) -o $(OUTPUT)/$(BINARY)-$@

openbsd-amd64:
	GOARCH=amd64 GOOS=openbsd $(GO_BUILD) -o $(OUTPUT)/$(BINARY)-$@

openbsd-arm64:
	GOARCH=arm64 GOOS=openbsd $(GO_BUILD) -o $(OUTPUT)/$(BINARY)-$@

windows-amd64:
	GOARCH=amd64 GOOS=windows $(GO_BUILD) -o $(OUTPUT)/$(BINARY)-$@.exe

gz_releases=$(addsuffix .gz, $(PLATFORM_LIST))
zip_releases=$(addsuffix .zip, $(WINDOWS_ARCH_LIST))

$(gz_releases): %.gz : %
	chmod +x $(OUTPUT)/$(BINARY)-$(basename $@)
	gzip -f -S .gz $(OUTPUT)/$(BINARY)-$(basename $@)

$(zip_releases): %.zip : %
	zip -m -j $(OUTPUT)/$(BINARY)-$(basename $@).zip $(OUTPUT)/$(BINARY)-$(basename $@).exe

all-arch: $(PLATFORM_LIST) $(WINDOWS_ARCH_LIST)

releases: $(gz_releases) $(zip_releases)

clean:
	rm $(OUTPUT)/*
