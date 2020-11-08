DIR = "bin"
NAME = "tun2socks"

TAGS = ""
BUILD_FLAGS = "-v"

VERSION = $(shell git describe --tags || echo "unknown version")
BUILD_TIME = $(shell date -u '+%FT%TZ')

LDFLAGS += -w -s -buildid=
LDFLAGS += -X "main.Version=$(VERSION)"
LDFLAGS += -X "main.BuildTime=$(BUILD_TIME)"  # RFC3339
GO_BUILD = CGO_ENABLED=0 go build $(BUILD_FLAGS) -ldflags '$(LDFLAGS)' -tags '$(TAGS)' -trimpath

PLATFORM_LIST = \
	darwin-amd64 \
	freebsd-amd64 \
	freebsd-arm64 \
	linux-amd64 \
	linux-arm64 \
	openbsd-amd64 \
	openbsd-arm64 \

.PHONY: all docker $(PLATFORM_LIST)

all: $(PLATFORM_LIST)

docker:
	$(GO_BUILD) -o $(DIR)/$(NAME)-$@

darwin-amd64:
	GOARCH=amd64 GOOS=darwin $(GO_BUILD) -o $(DIR)/$(NAME)-$@

freebsd-amd64:
	GOARCH=amd64 GOOS=freebsd $(GO_BUILD) -o $(DIR)/$(NAME)-$@

freebsd-arm64:
	GOARCH=arm64 GOOS=freebsd $(GO_BUILD) -o $(DIR)/$(NAME)-$@

linux-amd64:
	GOARCH=amd64 GOOS=linux $(GO_BUILD) -o $(DIR)/$(NAME)-$@

linux-arm64:
	GOARCH=arm64 GOOS=linux $(GO_BUILD) -o $(DIR)/$(NAME)-$@

openbsd-amd64:
	GOARCH=amd64 GOOS=openbsd $(GO_BUILD) -o $(DIR)/$(NAME)-$@

openbsd-arm64:
	GOARCH=arm64 GOOS=openbsd $(GO_BUILD) -o $(DIR)/$(NAME)-$@

zip_releases=$(addsuffix .zip, $(PLATFORM_LIST))

$(zip_releases): %.zip : %
	zip -m -j $(DIR)/$(NAME)-$(basename $@).zip $(DIR)/$(NAME)-$(basename $@)

releases: $(zip_releases)

clean:
	rm $(DIR)/*
