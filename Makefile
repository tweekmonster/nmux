DIST = $(PWD)/dist
GO_VERSION = 1.7.1
SOURCES = $(shell find . \( -path './.git*' -o -path './vendor' \) \
					-prune -o -type f -name '*.go' -print)
XGO_TARGETS := linux/amd64 darwin/amd64 windows/amd64
XGO_BUILD_TARGETS := $(foreach t,$(XGO_TARGETS),$(DIST)/$(shell echo "$(t)" \
	| sed 's!\(.*\)/\(.*\)!\2/\1!')/nmux)
XGO_BUILD_TARGETS := $(foreach t,$(XGO_BUILD_TARGETS), \
	$(shell echo "$(t)" | sed 's!.*windows.*!&.exe!'))
ADDR := "127.0.0.1:9999"

.PHONY: all clean generate run-server run-gui

all: $(DIST) generate $(XGO_BUILD_TARGETS)

clean:
	rm -rf $(DIST)

run-gui: $(DIST)
	go build -o $(DIST)/nmux-client ./cmd/nmux
	-$(DIST)/nmux-client --addr="$(ADDR)"
	rm $(DIST)/nmux-client

run-server: $(DIST)
	go build -o $(DIST)/nmux-server ./cmd/nmux
	@# Running from /tmp until I can figure out why prompts causes the API to
	@# become unresponsive.
	-cd /tmp && $(DIST)/nmux-server --server --addr="$(ADDR)"
	rm $(DIST)/nmux-server

generate:
	go generate

$(XGO_BUILD_TARGETS): $(SOURCES)
	$(eval t := $(wordlist 2,3,$(subst /, ,$@)))
	$(eval target := $(word 2,$(t))/$(word 1,$(t)))
	mkdir -p "$(@D)"
	mkdir -p --mode=2775 "$(@D)/tmp"
	xgo -go $(GO_VERSION) --targets="$(target)" -dest "$(@D)/tmp" ./cmd/nmux
	cp "$(@D)/tmp/nmux"* "$@"
	rm -rf "$(@D)/tmp"
	touch "$@"

$(DIST):
	mkdir -p $@
