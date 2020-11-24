SHELL=/usr/bin/env bash

all: build
.PHONY: all

unexport GOFLAGS

GOVERSION:=$(shell go version | cut -d' ' -f 3 | cut -d. -f 2)
ifeq ($(shell expr $(GOVERSION) \< 14), 1)
$(warning Your Golang version is go 1.$(GOVERSION))
$(error Update Golang to version $(shell grep '^go' go.mod))
endif

# git modules that need to be loaded
MODULES:=

CLEAN:=
BINS:=

ldflags=-X=github.com/guoxiaopeng875/lotus-adapter/build.CurrentCommit=+git.$(subst -,.,$(shell git describe --always --match=NeVeRmAtCh --dirty 2>/dev/null || git rev-parse --short HEAD 2>/dev/null))
ifneq ($(strip $(LDFLAGS)),)
	ldflags+=-extldflags=$(LDFLAGS)
endif

GOFLAGS+=-ldflags="$(ldflags)"


## FFI

FFI_PATH:=extern/filecoin-ffi/
FFI_DEPS:=.install-filcrypto
FFI_DEPS:=$(addprefix $(FFI_PATH),$(FFI_DEPS))

$(FFI_DEPS): build/.filecoin-install ;

build/.filecoin-install: $(FFI_PATH)
	$(MAKE) -C $(FFI_PATH) $(FFI_DEPS:$(FFI_PATH)%=%)
	@touch $@

MODULES+=$(FFI_PATH)
BUILD_DEPS+=build/.filecoin-install
CLEAN+=build/.filecoin-install

$(MODULES): build/.update-modules ;

# dummy file that marks the last time modules were updated
build/.update-modules:
	git submodule update --init --recursive
	touch $@

# end git modules

## MAIN BINARIES

CLEAN+=build/.update-modules

deps: $(BUILD_DEPS)
.PHONY: deps

clean:
	rm -rf $(CLEAN) $(BINS)
	-$(MAKE) -C $(FFI_PATH) clean
.PHONY: clean

dist-clean:
	git clean -xdff
	git submodule deinit --all -f
.PHONY: dist-clean

2k: GOFLAGS+=-tags=2k
2k: lotus-gateway

lotus-gateway: $(BUILD_DEPS)
	rm -f lotus-gateway
	go build $(GOFLAGS) -o lotus-gateway ./cmd/lotus-gateway
.PHONY: lotus-gateway
BINS+=lotus-gateway

lotus-monitor: $(BUILD_DEPS)
	rm -f lotus-monitor
	go build $(GOFLAGS) -o lotus-monitor ./cmd/lotus-monitor
.PHONY: lotus-monitor
BINS+=lotus-monitor

lotus-chainwatch:
	rm -f lotus-chainwatch
	go build $(GOFLAGS) -o lotus-chainwatch ./cmd/lotus-chainwatch
.PHONY: lotus-chainwatch
BINS+=lotus-chainwatch
