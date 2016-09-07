package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/zserge/hid"
)

func shell(device hid.Device) {
	if err := device.Open(); err != nil {
		log.Println("Open error: ", err)
		return
	}
	defer device.Close()

	if report, err := device.HIDReport(); err != nil {
		log.Println("HID report error:", err)
		return
	} else {
		log.Println("HID report", hex.EncodeToString(report))
	}

	go func() {
		for {
			if buf, err := device.Read(-1, 1*time.Second); err == nil {
				log.Println("Input report:  ", hex.EncodeToString(buf))
			}
		}
	}()

	commands := map[string]func([]byte){
		"output": func(b []byte) {
			if len(b) == 0 {
				log.Println("Invalid input: output report data expected")
			} else if n, err := device.Write(b, 1*time.Second); err != nil {
				log.Println("Output report write failed:", err)
			} else {
				log.Printf("Output report: written %d bytes\n", n)
			}
		},
		"set-feature": func(b []byte) {
			if len(b) == 0 {
				log.Println("Invalid input: feature report data expected")
			} else if err := device.SetReport(0, b); err != nil {
				log.Println("Feature report write failed:", err)
			} else {
				log.Printf("Feature report: " + hex.EncodeToString(b) + "\n")
			}
		},
		"get-feature": func(b []byte) {
			if b, err := device.GetReport(0); err != nil {
				log.Println("Feature report read failed:", err)
			} else {
				log.Println("Feature report: " + hex.EncodeToString(b) + "\n")
			}
		},
	}

	var completer = readline.NewPrefixCompleter(
		readline.PcItem("output"),
		readline.PcItem("set-feature"),
		readline.PcItem("get-feature"),
	)
	rl, err := readline.NewEx(&readline.Config{
		Prompt:       "> ",
		AutoComplete: completer,
	})
	if err != nil {
		panic(err)
	}

	defer rl.Close()
	log.SetOutput(rl.Stderr())

out:
	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}
		line = strings.ToLower(line)
		for cmd, f := range commands {
			if strings.HasPrefix(line, cmd) {
				line = strings.TrimSpace(line[len(cmd):])
				raw := []byte{}
				if len(line) > 0 {
					raw = make([]byte, len(line)/2, len(line)/2)
					if _, err := hex.Decode(raw, []byte(line)); err != nil {
						log.Println("Invalid input:", err)
						log.Println(">>", hex.EncodeToString(raw))
						continue out
					}
				}
				f(raw)
				continue out
			}
		}
	}
}

func main() {

	if len(os.Args) == 2 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Println("USAGE:")
		fmt.Printf("  %s              list USB HID devices\n", os.Args[0])
		fmt.Printf("  %s <id>         open USB HID device shell for the given input report size\n", os.Args[0])
		fmt.Printf("  %s -h|--help    show this help\n", os.Args[0])
		fmt.Println()
		return
	}

	// Without arguments - enumerate all HID devices
	if len(os.Args) == 1 {
		found := false
		hid.UsbWalk(func(device hid.Device) {
			info := device.Info()
			fmt.Printf("%04x:%04x:%04x:%02x\n", info.Vendor, info.Product, info.Revision, info.Interface)
			found = true
		})
		if !found {
			fmt.Println("No USB HID devices found\n")
		}
		return
	}

	hid.UsbWalk(func(device hid.Device) {
		info := device.Info()
		id := fmt.Sprintf("%04x:%04x:%04x:%02x", info.Vendor, info.Product, info.Revision, info.Interface)
		if id != os.Args[1] {
			return
		}

		shell(device)
	})
}
