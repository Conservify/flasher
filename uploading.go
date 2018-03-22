package main

import (
	"fmt"
	"log"
	"path"
	"runtime"
	"sort"
	"strings"
	"time"

	"go.bug.st/serial.v1"
)

type UploadOptions struct {
	Boards    *PropertiesMap
	Platform  *PropertiesMap
	Board     string
	Binary    string
	Port      string
	Tools     string
	SkipTouch bool
}

func getPortsMap() map[string]bool {
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}
	m := make(map[string]bool)
	for _, p := range ports {
		m[p] = true
	}
	return m
}

func toKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func diffPortMaps(before map[string]bool, after map[string]bool) (added []string, removed []string) {
	added = make([]string, 0)
	removed = make([]string, 0)

	for p, _ := range after {
		if _, ok := before[p]; !ok {
			added = append(added, p)
		}
	}

	for p, _ := range before {
		if _, ok := after[p]; !ok {
			removed = append(removed, p)
		}
	}

	return
}

type PortDiscoverer struct {
	before map[string]bool
}

func NewPortDiscoveror() *PortDiscoverer {
	return &PortDiscoverer{
		before: getPortsMap(),
	}
}

func (pd *PortDiscoverer) discover() string {
	s := time.Now()

	for {
		after := getPortsMap()

		added, removed := diffPortMaps(pd.before, after)

		log.Printf("%v -> %v | %v %v\n", toKeys(pd.before), toKeys(after), removed, added)

		if len(added) > 0 {
			return added[0]
		}

		time.Sleep(500 * time.Millisecond)

		if time.Since(s) > 10*time.Second {
			break
		}

		pd.before = after
	}

	return ""
}

func listPorts() {
	ports := getPortsMap()
	log.Printf("Ports: %v", toKeys(ports))
}

func getPlatformKey() string {
	if runtime.GOOS == "darwin" {
		return "macosx"
	}
	return runtime.GOOS
}

func Upload(options *UploadOptions) error {
	board := options.Boards.ToSubtree(options.Board)
	tools := options.Platform.ToSubtree("tools")
	tool, _ := board.Lookup("upload.tool", make(map[string]string))
	u := board.Merge(tools.ToSubtree(tool))

	commandKey := "cmd." + getPlatformKey()
	platformSpecificCommand := u.Properties[commandKey]
	if platformSpecificCommand != "" {
		// Eventually we should probably be downoading the arch specific
		// package and reading the configurations there.
		if runtime.GOARCH == "arm" {
			u.Properties["cmd"] = platformSpecificCommand + "_arm"
		} else {
			u.Properties["cmd"] = platformSpecificCommand
		}
		log.Printf("Using platform specific upload command (tried %s): %s (%s %s)", commandKey, platformSpecificCommand, runtime.GOOS, runtime.GOARCH)
	} else {
		log.Printf("No platform specific upload command, (tried %s) using %s (%s %s)", commandKey, u.Properties["cmd"], runtime.GOOS, runtime.GOARCH)
	}

	port := options.Port
	if port == "" {
		port = NewPortDiscoveror().discover()
		if port == "" {
			return fmt.Errorf("No port")
		}
	} else {
		use1200bpsTouch := board.ToBool("upload.use_1200bps_touch")

		if !options.SkipTouch && use1200bpsTouch {
			log.Printf("Performing 1200bps touch...")

			mode := &serial.Mode{
				BaudRate: 1200,
			}
			p, err := serial.Open(port, mode)
			if err != nil {
				listPorts()
				log.Fatalf("Error: Touch failed on (%s): %v", port, err)
			}
			p.SetDTR(false)
			p.SetRTS(true)
			p.Close()

			port = NewPortDiscoveror().discover()
			if port == "" {
				if options.Port == "" {
					return fmt.Errorf("No port")
				}
				port = options.Port
			}
		}
	}

	log.Printf("Using port %s\n", port)

	u.Properties["upload.verbose"] = u.Properties["upload.params.verbose"]
	u.Properties["upload.verify"] = u.Properties["upload.params.verify"]
	u.Properties["runtime.tools.bossac-1.6.1-arduino.path"] = options.Tools
	u.Properties["serial.port.file"] = path.Base(port)
	u.Properties["build.path"] = path.Dir(options.Binary)
	u.Properties["build.project_name"] = strings.Replace(path.Base(options.Binary), path.Ext(options.Binary), "", -1)

	line, _ := options.Platform.Lookup(fmt.Sprintf("tools.%s.upload.pattern", tool), u.Properties)

	log.Printf(line)

	if err := ExecuteAndPipeCommandLine(line, "upload | "); err != nil {
		return err
	}

	return nil
}
