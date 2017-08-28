flasher: flasher.go
	go build -o flasher *.go

install: flasher
	mkdir -p ~/tools/lib
	cp -ar tools ~/tools/lib/flasher
	cp flasher ~/tools/bin
