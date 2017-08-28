package main

import (
	"flag"
	"fmt"
	"go.bug.st/serial.v1"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"
)

type configuration struct {
	Board     string
	Port      string
	Binary    string
	Tools     string
	SkipTouch bool
	Tail      bool
}

func searchForTools(config *configuration) string {
	if config.Tools != "" {
		return config.Tools
	}

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	candidates := []string{
		filepath.Join(dir, "tools"),
		filepath.Join(filepath.Dir(dir), "lib/flasher"),
	}

	for _, p := range candidates {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			return p
		}
	}

	return "./tools"
}

func main() {
	config := configuration{}

	flag.StringVar(&config.Board, "board", "adafruit_feather_m0", "board to upload to")
	flag.StringVar(&config.Port, "port", "", "port to upload to")
	flag.StringVar(&config.Binary, "binary", "", "path to the binary (required)")
	flag.StringVar(&config.Tools, "tools", "", "path to the tools directory")
	flag.BoolVar(&config.SkipTouch, "skip-touch", false, "skip the touch")
	flag.BoolVar(&config.Tail, "tail", false, "show serial")
	flag.Parse()

	if config.Binary == "" {
		flag.Usage()
		os.Exit(1)
	}

	config.Tools = searchForTools(&config)

	boardsPath := path.Join(config.Tools, "boards.txt")
	boards, err := NewPropertiesMapFromFile(boardsPath)
	if err != nil {
		log.Fatalf("Unable to open %s (%v)", boardsPath, err)
	}

	platformPath := path.Join(config.Tools, "platform.txt")
	platform, err := NewPropertiesMapFromFile(platformPath)
	if err != nil {
		log.Fatalf("Unable to open %s (%v)", platformPath, err)
	}

	Upload(&UploadOptions{
		Boards:    boards,
		Platform:  platform,
		SkipTouch: config.SkipTouch,
		Board:     config.Board,
		Port:      config.Port,
		Binary:    config.Binary,
		Tools:     config.Tools,
	})

	if config.Tail {
		time.Sleep(1 * time.Second)

		mode := &serial.Mode{
			BaudRate: 115200,
		}
		port, err := serial.Open(config.Port, mode)
		if err != nil {
			log.Fatalf("Unable to open %s (%v)", config.Port, err)
		}

		fmt.Println()

		buff := make([]byte, 100)
		for {
			n, err := port.Read(buff)
			if err != nil {
				log.Fatal(err)
				break
			}
			if n == 0 {
				break
			}
			fmt.Printf("%v", string(buff[:n]))
		}
	}
}
