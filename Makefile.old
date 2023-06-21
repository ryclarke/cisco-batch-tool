GO_SOURCES = $(shell find * -type f -name "*.go")
TARGET_OS = $(shell echo '$@' | sed 's|.*bin/\([^/]\+\)/.*|\1|')

TARGET_LINUX   := bin/linux/batch-tool
TARGET_DARWIN  := bin/darwin/batch-tool
TARGET_WINDOWS := bin/windows/batch-tool.exe
TARGETS := $(TARGET_LINUX) $(TARGET_DARWIN) $(TARGET_WINDOWS)

RELEASE_LINUX   := release/batch-tool-linux.txz
RELEASE_DARWIN  := release/batch-tool-darwin.txz
RELEASE_WINDOWS := release/batch-tool-windows.zip
RELEASES := $(RELEASE_LINUX) $(RELEASE_DARWIN) $(RELEASE_WINDOWS)

TARCMD = tar --transform='s|.*/||' caf $@ $^
ZIPCMD = zip -j $@ $^

.PHONY: build clean package

build: $(TARGETS)
$(TARGETS): $(GO_SOURCES)
	GOARCH=amd64 GOOS=$(TARGET_OS) go build -o $@

clean:
	rm -rf bin/ release/

package: $(RELEASES)
$(RELEASE_LINUX): $(TARGET_LINUX)
	@mkdir -p release/
	$(TARCMD)

$(RELEASE_DARWIN): $(TARGET_DARWIN)
	@mkdir -p release/
	$(TARCMD)

$(RELEASE_WINDOWS): $(TARGET_WINDOWS)
	@mkdir -p release/
	$(ZIPCMD)
