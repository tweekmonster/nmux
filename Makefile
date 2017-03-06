DIST := dist
GO_VERSION := 1.8
SOURCES := \
	$(shell find . \( -path './.git*' -o -path './vendor' -o -path '$(DIST)' \) \
	-prune -o -type f -name '*.go' -print)
BINDATA = $(shell find data -type f)
ADDR := "127.0.0.1:9999"

# args: BUILD_BIN PLATFORM ARCHITECTURE ARCHIVE_SUFFIX
define build_target

all: $(1) $(DIST)/$(2)-$(3).$(4)
archive: $(DIST)/$(2)-$(3).$(4)
.PHONY: $(2)

$(2): $(1)

$(1): $$(SOURCES)
	@scripts/build.sh $$(DIST) $(2) $(3)

$(DIST)/$(2)-$(3).$(4): $(1)
	@scripts/archive.sh $$< $$@

endef

.PHONY: all clean archive .generated run-server run-gui

all: $(DIST) .generated

clean:
	rm -rf $(DIST)
	rm -f bindata.go screen/const_string.go

$(DIST):
	mkdir -p $@

run-gui: $(DIST)
	go build -o $(DIST)/nmux-client ./cmd/nmux
	-$(DIST)/nmux-client --addr="$(ADDR)"
	rm $(DIST)/nmux-client

run-server: $(DIST) .generated $(SOURCES)
	go build -o $(DIST)/nmux-server ./cmd/nmux
	@# Running from /tmp until I can figure out why prompts causes the API to
	@# become unresponsive.
	@# @-cd /tmp && $(DIST)/nmux-server --server --addr="$(ADDR)"
	-$(DIST)/nmux-server --server --addr="$(ADDR)"
	rm $(DIST)/nmux-server

.generated: bindata.go screen/const_string.go

bindata.go: $(BINDATA)
	go-bindata ${BIN_DATA_ARGS} -pkg=nmux -prefix data/ data/...

screen/const_string.go: screen/const.go
	stringer -type=Op,Attr,Mode -output screen/const_string.go screen/

$(eval $(call build_target,$(DIST)/darwin-10.8/amd64/nmux.app/Contents/MacOS/nmux,darwin-10.8,amd64,tar.bz2))
# $(eval $(call build_target,$(DIST)/linux/amd64/nmux,linux,amd64,tar.bz2))
# $(eval $(call build_target,$(DIST)/windows/amd64/nmux.exe,windows,amd64,zip))
# $(eval $(call build_target,$(DIST)/windows/386/nmux.exe,windows,386,zip))
