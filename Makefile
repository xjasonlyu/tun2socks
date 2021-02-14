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

PLATFORM_LIST = \
	darwin-amd64 \
	freebsd-amd64 \
	freebsd-arm64 \
	linux-amd64 \
	linux-arm64 \
	openbsd-amd64 \
	openbsd-arm64 \
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

openbsd-amd64:
	GOARCH=amd64 GOOS=openbsd $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

openbsd-arm64:
	GOARCH=arm64 GOOS=openbsd $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@

windows-amd64:
	GOARCH=amd64 GOOS=windows $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY)-$@.exe

zip_releases := $(addsuffix .zip, $(PLATFORM_LIST))

$(zip_releases): %.zip: %
	@zip -m -j $(BUILD_DIR)/$(BINARY)-$(basename $@).zip $(BUILD_DIR)/$(BINARY)-$(basename $@)*

all-arch: $(PLATFORM_LIST)

releases: $(zip_releases)

clean:
	rm -rf $(BUILD_DIR)
