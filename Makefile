flasher: *.go
	go build -o flasher *.go

install: flasher
	mkdir -p ~/tools/lib
	cp -a tools ~/tools/lib/flasher
	sudo cp flasher ~/tools/bin
	sudo chown root. ~/tools/bin/flasher
	sudo chmod ug+s ~/tools/bin/flasher

clean:
	rm -f flasher
