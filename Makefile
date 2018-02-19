GOARCH ?= amd64
GO ?= env GOOS=linux GOARCH=$(GOARCH) go
UNAME := $(shell uname)
BUILD ?= build

$(BUILD)/flasher: *.go
	$(GO) build -o $(BUILD)/flasher *.go

$(BUILD):
	mkdir -p $(BUILD)

install: $(BUILD)/flasher
	echo $(UNAME)
	mkdir -p ~/tools/lib
	cp -a tools ~/tools/lib/flasher
	sudo cp $(BUILD)/flasher ~/tools/bin
ifeq ($(UNAME),Linux)
	sudo chown root. ~/tools/bin/flasher
	sudo chmod ug+s ~/tools/bin/flasher
endif

clean:
	rm -rf $(BUILD)
