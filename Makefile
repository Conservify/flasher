GOARCH ?= amd64
GO ?= env GOOS=linux GOARCH=$(GOARCH) go
UNAME := $(shell uname)

flasher: *.go
	$(GO) build -o flasher *.go

install: flasher
	echo $(UNAME)
	mkdir -p ~/tools/lib
	cp -a tools ~/tools/lib/flasher
	sudo cp flasher ~/tools/bin
ifeq ($(UNAME),Linux)
	sudo chown root. ~/tools/bin/flasher
	sudo chmod ug+s ~/tools/bin/flasher
endif

clean:
	rm -f flasher
