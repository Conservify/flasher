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

$(BUILDARCH)/flasher: $(BUILDARCH) *.go
	$(GO) build -o $(BUILDARCH)/flasher *.go

$(BUILDARCH):
	mkdir -p $(BUILDARCH)

install: $(BUILDARCH)/flasher
	echo $(UNAME)
	mkdir -p ~/tools/lib/flasher
	cp -a tools/* ~/tools/lib/flasher
	sudo cp $(BUILDARCH)/flasher ~/tools/bin
ifeq ($(UNAME),Linux)
	sudo chown root. ~/tools/bin/flasher
	sudo chmod ug+s ~/tools/bin/flasher
endif

clean:
	rm -rf $(BUILD)
