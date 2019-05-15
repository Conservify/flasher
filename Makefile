GOARCH ?= amd64
GOOS ?= linux
GO ?= env GOOS=$(GOOS) GOARCH=$(GOARCH) go
UNAME := $(shell uname)
BUILD ?= build
BUILDARCH ?= $(BUILD)/$(GOOS)-$(GOARCH)

all:
	GOOS=linux GOARCH=amd64 make binaries-all
	GOOS=linux GOARCH=arm make binaries-all
	GOOS=darwin GOARCH=amd64 make binaries-all

binaries-all: $(BUILDARCH)/flasher
	cd $(BUILDARCH) && zip -r ../$(GOOS)-$(GOARCH).zip *

$(BUILDARCH)/flasher: $(BUILDARCH) *.go
	$(GO) build -o $(BUILDARCH)/flasher *.go

$(BUILDARCH):
	mkdir -p $(BUILDARCH)
	cp -a tools $(BUILDARCH)

install: all
	echo $(UNAME)
ifeq ($(UNAME),Linux)
	rm -rf ~/tools/lib/flasher
	mkdir -p ~/tools/lib
	cp -a tools ~/tools/lib/flasher
	sudo rm ~/tools/bin/flasher
	sudo cp $(BUILD)/linux-amd64/flasher ~/tools/bin
	sudo chown root. ~/tools/bin/flasher
	sudo chmod ug+s ~/tools/bin/flasher
endif
ifeq ($(UNAME),Darwin)
	rm -rf ~/tools/lib/flasher
	mkdir -p ~/tools/lib
	cp -a tools ~/tools/lib/flasher
	sudo rm ~/tools/bin/flasher
	sudo cp $(BUILD)/darwin-amd64/flasher ~/tools/bin
endif

clean:
	rm -rf $(BUILD)

deps:
	$(GO) get go.bug.st/serial.v1

.PHONY: $(BUILD)
