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
	Board          string
	Port           string
	Binary         string
	Tools          string
	SkipTouch      bool
	Tail           bool
	TailInactivity int
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

func echoSerial(config *configuration, c *chan bool) {
	mode := &serial.Mode{
		BaudRate: 115200,
	}
	port, err := serial.Open(config.Port, mode)
	if err != nil {
		log.Fatalf("Unable to open %s (%v)", config.Port, err)
	}

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
		*c <- true
		fmt.Printf("%v", string(buff[:n]))
	}
}

func main() {
	config := configuration{}

	flag.StringVar(&config.Board, "board", "adafruit_feather_m0", "board to upload to")
	flag.StringVar(&config.Port, "port", "", "port to upload to")
	flag.StringVar(&config.Binary, "binary", "", "path to the binary (required)")
	flag.StringVar(&config.Tools, "tools", "", "path to the tools directory")
	flag.BoolVar(&config.SkipTouch, "skip-touch", false, "skip the touch")
	flag.BoolVar(&config.Tail, "tail", false, "show serial")
	flag.IntVar(&config.TailInactivity, "tail-inactivity", 0, "inactive time until quitting tail")
	flag.Parse()

	if config.Binary == "" && !config.Tail {
		flag.Usage()
		os.Exit(2)
	}

	if config.Binary != "" {
		if _, err := os.Stat(config.Binary); os.IsNotExist(err) {
			log.Fatalf("No such binary '%s'", config.Binary)
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
	}

	if config.Tail {
		time.Sleep(1 * time.Second)

		ch := make(chan bool)
		go echoSerial(&config, &ch)

		go func() {
			for {
				time.Sleep(1 * time.Second)
				ch <- false
			}
		}()

		previous := time.Now()
		for {
			data := <-ch
			if data {
				previous = time.Now()
			}

			if config.TailInactivity > 0 {
				ago := time.Duration(-config.TailInactivity) * time.Second
				to := time.Now().Add(ago)
				if previous.Before(to) {
					break
				}
			}
		}
	}
}
