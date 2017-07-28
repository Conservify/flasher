package main

import (
	"flag"
	"log"
	"os"
	"path"
)

type configuration struct {
	Board  string
	Port   string
	Binary string
	Tools  string
}

func main() {
	config := configuration{}

	flag.StringVar(&config.Board, "board", "adafruit_feather_m0", "board to upload to")
	flag.StringVar(&config.Port, "port", "", "port to upload to")
	flag.StringVar(&config.Binary, "binary", "", "path to the binary (required)")
	flag.StringVar(&config.Tools, "tools", "./tools", "path to the tools directory")
	flag.Parse()

	if config.Binary == "" {
		flag.Usage()
		os.Exit(1)
	}

	boards, err := NewPropertiesMapFromFile(path.Join(config.Tools, "boards.txt"))
	if err != nil {
		log.Fatal(err)
	}

	platform, err := NewPropertiesMapFromFile(path.Join(config.Tools, "platform.txt"))
	if err != nil {
		log.Fatal(err)
	}

	Upload(&UploadOptions{
		Boards:   boards,
		Platform: platform,
		Board:    config.Board,
		Port:     config.Port,
		Binary:   config.Binary,
	})

	/*
		mode := &serial.Mode{
			BaudRate: 115200,
		}
		port, err := serial.Open("/dev/ttyUSB0", mode)
		if err != nil {
			log.Fatal(err)
		}
	*/
}
