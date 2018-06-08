GOARCH ?= amd64
GOOS ?= linux
GO ?= env GOOS=$(GOOS) GOARCH=$(GOARCH) go
BUILD ?= build
BUILDARCH ?= $(BUILD)/$(GOOS)-$(GOARCH)
UNAME = $(shell uname)

all:
	GOOS=linux GOARCH=amd64 make binaries-all
	GOOS=linux GOARCH=arm make binaries-all
	GOOS=darwin GOARCH=amd64 make binaries-all

binaries-all: $(BUILDARCH)/flasher

$(BUILDARCH)/flasher: *.go
	$(GO) build -o $(BUILDARCH)/flasher *.go

$(BUILD):
	mkdir -p $(BUILD)

install: all
	echo $(UNAME)
ifeq ($(UNAME),Linux)
	mkdir -p ~/tools/lib
	cp -a tools ~/tools/lib/flasher
	sudo cp $(BUILD)/linux-amd64/flasher ~/tools/bin
	sudo chown root. ~/tools/bin/flasher
	sudo chmod ug+s ~/tools/bin/flasher
endif
ifeq ($(UNAME),Darwin)
	mkdir -p ~/tools/lib
	cp -a tools ~/tools/lib/flasher
	sudo cp $(BUILD)/darwin-amd64/flasher ~/tools/bin
endif

clean:
	rm -rf $(BUILD)

.PHONY: $(BUILD)
