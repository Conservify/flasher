package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"go.bug.st/serial.v1"

	tooling "github.com/conservify/tooling"
)

type configuration struct {
	Board  string
	Port   string
	Binary string
	Tools  string

	SkipTouch     bool
	Touch         bool
	Verbose       bool
	Verify        bool
	UploadQuietly bool

	Tail            bool
	TailAppend      string
	TailInactivity  int
	TailReopen      bool
	TailTriggerFail string
	TailTriggerPass string
	TailTriggerStop string

	FlashOffset int
}

func openSerial(config *configuration) (serial.Port, error) {
	mode := &serial.Mode{
		BaudRate: 115200,
	}
	port, err := serial.Open(config.Port, mode)
	if err != nil {
		return nil, err
	}

	return port, nil
}

type EchoStatus struct {
	Data   bool
	Exited bool
}

func openFile(config *configuration) *os.File {
	if config.TailAppend == "" {
		return nil
	}

	log.Printf("Logging to %s...", config.TailAppend)

	file, err := os.OpenFile(config.TailAppend, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Unable to open %s: %v", config.TailAppend, err)
	}

	return file
}

type StopTriggers struct {
	Stop *regexp.Regexp
	Pass *regexp.Regexp
	Fail *regexp.Regexp
}

func NewStopTriggers(config *configuration) (st *StopTriggers) {
	var stop *regexp.Regexp
	var pass *regexp.Regexp
	var fail *regexp.Regexp

	if config.TailTriggerStop != "" {
		stop = regexp.MustCompile(config.TailTriggerStop)
		log.Printf("Stopping on '%s'", config.TailTriggerStop)
	}

	if config.TailTriggerPass != "" {
		pass = regexp.MustCompile(config.TailTriggerPass)
		log.Printf("Passing on '%s'", config.TailTriggerPass)
	}

	if config.TailTriggerFail != "" {
		fail = regexp.MustCompile(config.TailTriggerFail)
		log.Printf("Failing on '%s'", config.TailTriggerFail)
	}

	return &StopTriggers{
		Stop: stop,
		Pass: pass,
		Fail: fail,
	}
}

func (st *StopTriggers) Apply(line string) (exitCode int) {
	if st.Pass != nil {
		if st.Pass.MatchString(line) {
			return 0
		}

	}
	if st.Fail != nil {
		if st.Fail.MatchString(line) {
			return 2
		}
	}
	if st.Stop != nil {
		if st.Stop.MatchString(line) {
			return 0
		}
	}
	return -1
}

func echoSerial(config *configuration, port serial.Port, c chan *EchoStatus) {
	defer port.Close()

	var w *bufio.Writer
	file := openFile(config)
	if file != nil {
		w = bufio.NewWriter(file)

		defer file.Close()

		defer w.Flush()
	}

	buff := make([]byte, 256)

	triggers := NewStopTriggers(config)

	for {
		n, err := port.Read(buff)
		if err != nil {
			log.Printf("Error reading: %v", err)
			break
		}
		if n == 0 {
			break
		}
		c <- &EchoStatus{
			Data: true,
		}
		// This is probably controversial:
		sanitized := strings.Replace(string(buff[:n]), "\r", "", -1)
		if w != nil {
			fmt.Printf("%v", sanitized)
			w.WriteString(sanitized)
			w.Flush()
		} else {
			fmt.Printf("%v", sanitized)
		}

		exitCode := triggers.Apply(sanitized)
		if exitCode >= 0 {
			os.Exit(exitCode)
		}
	}

	c <- &EchoStatus{
		Exited: true,
	}
}

func main() {
	config := configuration{}

	flag.StringVar(&config.Tools, "tools", "", "path to the tools directory")
	flag.StringVar(&config.Board, "board", "adafruit_feather_m0", "board to upload to")

	flag.BoolVar(&config.SkipTouch, "skip-touch", false, "skip the touch")
	flag.BoolVar(&config.Touch, "touch", false, "touch")
	flag.BoolVar(&config.Verbose, "verbose", false, "verbose")
	flag.BoolVar(&config.Verify, "verify", false, "verify")

	flag.StringVar(&config.Port, "port", "", "port to upload to")
	flag.StringVar(&config.Binary, "binary", "", "path to the binary (required)")
	flag.IntVar(&config.FlashOffset, "flash-offset", 0, "flash offset to flash program")

	flag.BoolVar(&config.UploadQuietly, "upload-quietly", false, "hide upload progress")

	flag.BoolVar(&config.Tail, "tail", false, "show serial")
	flag.StringVar(&config.TailAppend, "append", "", "append tail to file")
	flag.IntVar(&config.TailInactivity, "tail-inactivity", 0, "inactive time until quitting tail")
	flag.BoolVar(&config.TailReopen, "tail-reopen", true, "tail again after inactivity or file loss")
	flag.StringVar(&config.TailTriggerStop, "tail-trigger-stop", "", "tail trigger stop")
	flag.StringVar(&config.TailTriggerPass, "tail-trigger-pass", "", "tail trigger pass")
	flag.StringVar(&config.TailTriggerFail, "tail-trigger-fail", "", "tail trigger fail")

	flag.Parse()

	if config.Binary == "" && !config.Tail {
		flag.Usage()
		os.Exit(2)
	}

	pd := tooling.NewPortDiscoveror()

	if config.Binary != "" {
		if config.FlashOffset == 0 {
			flag.Usage()
			os.Exit(2)
		}
		if _, err := os.Stat(config.Binary); os.IsNotExist(err) {
			log.Fatalf("Error: No such binary '%s'", config.Binary)
		}

		ae := tooling.NewArduinoEnvironment()

		err := ae.Locate(config.Tools)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		portPath, err := filepath.EvalSymlinks(config.Port)
		if err != nil {
			log.Fatalf("Unable to evaluate symlinks %s (%v)", config.Port, err)
		}

		tooling.Upload(&tooling.UploadOptions{
			Arduino:     ae,
			SkipTouch:   config.SkipTouch,
			Board:       config.Board,
			Port:        portPath,
			Binary:      config.Binary,
			FlashOffset: config.FlashOffset,
			Verbose:     config.Verbose,
			Verify:      config.Verify,
			Quietly:     config.UploadQuietly,
		})
	}

	if config.Tail {
		if config.Touch {
			if config.Port != "" {
				portPath, err := filepath.EvalSymlinks(config.Port)
				if err != nil {
					log.Fatalf("Unable to evaluate symlinks %s (%v)", config.Port, err)
				}

				tooling.Touch(portPath)
			}
		}

		if config.Touch || config.Binary != "" { // Did we upload?
			time.Sleep(500 * time.Millisecond)
		}

		for {
			if _, err := os.Stat(config.Port); os.IsNotExist(err) {
				log.Printf("Port '%s' disappeared, scanning...", config.Port)
				config.Port = pd.Discover()
				if config.Port == "" {
					log.Fatalf("Error: Unable to find port to tail.")
				}
			}

			ch := make(chan *EchoStatus)

			port, err := openSerial(&config)
			if err != nil {
				if config.TailReopen {
					time.Sleep(500 * time.Millisecond)
					continue
				} else {
					log.Fatalf("Error: Unable to open port: %v", err)
				}
			}

			go echoSerial(&config, port, ch)

			go func() {
				for {
					time.Sleep(1 * time.Second)
					ch <- &EchoStatus{}
				}
			}()

			previous := time.Now()

			for {
				status := <-ch
				if status.Data {
					previous = time.Now()
				}
				if status.Exited {
					port.Close()
					break
				}

				if config.TailInactivity > 0 {
					ago := time.Duration(-config.TailInactivity) * time.Second
					to := time.Now().Add(ago)
					if previous.Before(to) {
						port.Close()
						log.Printf("Tail inactive!")
						break
					}
				}
			}

			if !config.TailReopen {
				break
			}
		}
	}
}
