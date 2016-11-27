DIST = $(PWD)/dist
GO_VERSION = 1.7.1
SOURCES = $(shell find . \( -path './.git*' -o -path './vendor' \) \
					-prune -o -type f -name '*.go' -print)
XGO_TARGETS := linux/amd64 darwin/amd64 windows/amd64
XGO_BUILD_TARGETS := $(foreach t,$(XGO_TARGETS),$(DIST)/$(shell echo "$(t)" \
	| sed 's!\(.*\)/\(.*\)!\2/\1!')/nmux)
XGO_BUILD_TARGETS := $(foreach t,$(XGO_BUILD_TARGETS), \
	$(shell echo "$(t)" | sed 's!.*windows.*!&.exe!'))

.PHONY: all clean generate run-server run-gui

all: $(DIST) generate $(XGO_BUILD_TARGETS)

clean:
	rm -rf $(DIST)

quick-build: $(DIST)
	go build -o $(DIST)/nmux ./cmd/nmux

run-gui: quick-build
	$(DIST)/nmux

run-server: quick-build
	-cd /tmp && $(DIST)/nmux --server --addr="127.0.0.1:9999"
	rm $(DIST)/nmux

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
