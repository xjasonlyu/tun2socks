BINARY := tun2socks
MODULE := github.com/xjasonlyu/tun2socks

BUILD_DIR     := build
BUILD_TAGS    :=
BUILD_FLAGS   := -v
BUILD_COMMIT  := $(shell git describe --dirty --always)
BUILD_VERSION := $(shell git describe --abbrev=0 --tags HEAD)

CGO_ENABLED := 0
GO111MODULE := on

LDFLAGS += -w -s -buildid=
LDFLAGS += -X "$(MODULE)/constant.Version=$(BUILD_VERSION)"
LDFLAGS += -X "$(MODULE)/constant.GitCommit=$(BUILD_COMMIT)"

GO_BUILD = GO111MODULE=$(GO111MODULE) CGO_ENABLED=$(CGO_ENABLED) \
	go build $(BUILD_FLAGS) -ldflags '$(LDFLAGS)' -tags '$(BUILD_TAGS)' -trimpath

UNIX_ARCH_LIST = \
	darwin-amd64 \
	freebsd-amd64 \
	freebsd-arm64 \
	linux-amd64 \
	linux-arm64 \
	linux-mips64 \
	linux-mips64le \
	linux-ppc64 \
    linux-ppc64le \
	openbsd-amd64 \
	openbsd-arm64 \

WINDOWS_ARCH_LIST = \
	windows-amd64 \

all: linux-amd64 darwin-amd64 windows-amd64

tun2socks:
	$(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)

darwin-amd64:
	GOARCH=amd64 GOOS=darwin $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

freebsd-amd64:
	GOARCH=amd64 GOOS=freebsd $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

freebsd-arm64:
	GOARCH=arm64 GOOS=freebsd $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

linux-amd64:
	GOARCH=amd64 GOOS=linux $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

linux-arm64:
	GOARCH=arm64 GOOS=linux $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

linux-mips64:
	GOARCH=mips64 GOOS=linux $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

linux-mips64le:
	GOARCH=mips64le GOOS=linux $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

linux-ppc64:
	GOARCH=ppc64 GOOS=linux $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

linux-ppc64le:
	GOARCH=ppc64le GOOS=linux $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

openbsd-amd64:
	GOARCH=amd64 GOOS=openbsd $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

openbsd-arm64:
	GOARCH=arm64 GOOS=openbsd $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

windows-amd64:
	GOARCH=amd64 GOOS=windows $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@.exe

unix_releases := $(addsuffix .zip, $(UNIX_ARCH_LIST))
windows_releases := $(addsuffix .zip, $(WINDOWS_ARCH_LIST))

$(unix_releases): %.zip: %
	@zip -qmj $(BUILD_DIR)/$(BINARY)-$(basename $@).zip $(BUILD_DIR)/$(BINARY)-$(basename $@)

$(windows_releases): %.zip: %
	@zip -qmj $(BUILD_DIR)/$(BINARY)-$(basename $@).zip $(BUILD_DIR)/$(BINARY)-$(basename $@).exe

all-arch: $(UNIX_ARCH_LIST) $(WINDOWS_ARCH_LIST)

releases: $(unix_releases) $(windows_releases)

clean:
	rm -rf $(BUILD_DIR)
