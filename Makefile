BINARY := tun2socks
MODULE := github.com/xjasonlyu/tun2socks/v2

BUILD_DIR     := build
BUILD_TAGS    :=
BUILD_FLAGS   := -v
BUILD_COMMIT  := $(shell git rev-parse --short HEAD)
BUILD_VERSION := $(shell git describe --abbrev=0 --tags HEAD)

CGO_ENABLED := 0
GO111MODULE := on

LDFLAGS += -w -s -buildid=
LDFLAGS += -X "$(MODULE)/internal/version.Version=$(BUILD_VERSION)"
LDFLAGS += -X "$(MODULE)/internal/version.GitCommit=$(BUILD_COMMIT)"

GO_BUILD = GO111MODULE=$(GO111MODULE) CGO_ENABLED=$(CGO_ENABLED) \
	go build $(BUILD_FLAGS) -ldflags '$(LDFLAGS)' -tags '$(BUILD_TAGS)' -trimpath

UNIX_ARCH_LIST = \
	darwin-amd64 \
	darwin-amd64-v3 \
	darwin-arm64 \
	freebsd-386 \
	freebsd-amd64 \
	freebsd-amd64-v3 \
	freebsd-arm64 \
	linux-386 \
	linux-amd64 \
	linux-amd64-v3 \
	linux-arm64 \
	linux-armv5 \
	linux-armv6 \
	linux-armv7 \
	linux-mips-softfloat \
	linux-mips-hardfloat \
	linux-mipsle-softfloat \
	linux-mipsle-hardfloat \
	linux-mips64 \
	linux-mips64le \
	linux-ppc64 \
	linux-ppc64le \
	linux-s390x \
	linux-loong64 \
	openbsd-amd64 \
	openbsd-amd64-v3 \
	openbsd-arm64

WINDOWS_ARCH_LIST = \
	windows-386 \
	windows-amd64 \
	windows-amd64-v3 \
	windows-arm64 \
	windows-arm32v7

all: linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64

debug: BUILD_TAGS += debug
debug: all

tun2socks:
	$(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)

darwin-amd64:
	GOARCH=amd64 GOOS=darwin $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

darwin-amd64-v3:
	GOARCH=amd64 GOOS=darwin GOAMD64=v3 $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

darwin-arm64:
	GOARCH=arm64 GOOS=darwin $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

freebsd-386:
	GOARCH=386 GOOS=freebsd $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

freebsd-amd64:
	GOARCH=amd64 GOOS=freebsd $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

freebsd-amd64-v3:
	GOARCH=amd64 GOOS=freebsd GOAMD64=v3 $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

freebsd-arm64:
	GOARCH=arm64 GOOS=freebsd $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

linux-386:
	GOARCH=386 GOOS=linux $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

linux-amd64:
	GOARCH=amd64 GOOS=linux $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

linux-amd64-v3:
	GOARCH=amd64 GOOS=linux GOAMD64=v3 $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

linux-arm64:
	GOARCH=arm64 GOOS=linux $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

linux-armv5:
	GOARCH=arm GOARM=5 GOOS=linux $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

linux-armv6:
	GOARCH=arm GOARM=6 GOOS=linux $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

linux-armv7:
	GOARCH=arm GOARM=7 GOOS=linux $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

linux-mips-softfloat:
	GOARCH=mips GOMIPS=softfloat GOOS=linux $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

linux-mips-hardfloat:
	GOARCH=mips GOMIPS=hardfloat GOOS=linux $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

linux-mipsle-softfloat:
	GOARCH=mipsle GOMIPS=softfloat GOOS=linux $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

linux-mipsle-hardfloat:
	GOARCH=mipsle GOMIPS=hardfloat GOOS=linux $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

linux-mips64:
	GOARCH=mips64 GOOS=linux $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

linux-mips64le:
	GOARCH=mips64le GOOS=linux $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

linux-ppc64:
	GOARCH=ppc64 GOOS=linux $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

linux-ppc64le:
	GOARCH=ppc64le GOOS=linux $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

linux-s390x:
	GOARCH=s390x GOOS=linux $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

linux-loong64:
	GOARCH=loong64 GOOS=linux $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

openbsd-amd64:
	GOARCH=amd64 GOOS=openbsd $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

openbsd-amd64-v3:
	GOARCH=amd64 GOOS=openbsd GOAMD64=v3 $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

openbsd-arm64:
	GOARCH=arm64 GOOS=openbsd $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

windows-386:
	GOARCH=386 GOOS=windows $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@.exe

windows-amd64:
	GOARCH=amd64 GOOS=windows $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@.exe

windows-amd64-v3:
	GOARCH=amd64 GOOS=windows GOAMD64=v3 $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@.exe

windows-arm64:
	GOARCH=arm64 GOOS=windows $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@.exe

windows-arm32v7:
	GOARCH=arm GOARM=7 GOOS=windows $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@.exe

unix_releases := $(addsuffix .zip, $(UNIX_ARCH_LIST))
windows_releases := $(addsuffix .zip, $(WINDOWS_ARCH_LIST))

$(unix_releases): %.zip: %
	@zip -qmj $(BUILD_DIR)/$(BINARY)-$(basename $@).zip $(BUILD_DIR)/$(BINARY)-$(basename $@)

$(windows_releases): %.zip: %
	@zip -qmj $(BUILD_DIR)/$(BINARY)-$(basename $@).zip $(BUILD_DIR)/$(BINARY)-$(basename $@).exe

all-arch: $(UNIX_ARCH_LIST) $(WINDOWS_ARCH_LIST)

releases: $(unix_releases) $(windows_releases)

lint:
	GOOS=darwin  golangci-lint run ./...
	GOOS=windows golangci-lint run ./...
	GOOS=linux   golangci-lint run ./...
	GOOS=freebsd golangci-lint run ./...
	GOOS=openbsd golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR)
