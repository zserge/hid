package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/peterh/liner"
	"github.com/zserge/hid"
)

func shell(device hid.Device, inputReportSize int) {
	if err := device.Open(); err != nil {
		fmt.Println("Open error: ", err)
		return
	}
	defer device.Close()

	if report, err := device.HIDReport(); err != nil {
		fmt.Println("HID report error:", err)
		return
	} else {
		fmt.Println("HID report", hex.EncodeToString(report))
	}

	go func() {
		for {
			if buf, err := device.Read(inputReportSize, 1*time.Second); err == nil {
				fmt.Println("\rInput report:  ", hex.EncodeToString(buf))
				syscall.Kill(syscall.Getpid(), syscall.SIGWINCH)
			}
		}
	}()

	commands := map[string]func([]byte){
		"output": func(b []byte) {
			if len(b) == 0 {
				fmt.Println("Invalid input: output report data expected")
			} else if n, err := device.Write(b, 1*time.Second); err != nil {
				fmt.Println("Output report write failed:", err)
			} else {
				fmt.Printf("Output report: written %d bytes\n", n)
			}
		},
		"set-feature": func(b []byte) {
			if len(b) == 0 {
				fmt.Println("Invalid input: feature report data expected")
			} else if err := device.SetReport(0, b); err != nil {
				fmt.Println("Feature report write failed:", err)
			} else {
				fmt.Printf("Feature report: " + hex.EncodeToString(b) + "\n")
			}
		},
		"get-feature": func(b []byte) {
			if b, err := device.GetReport(0); err != nil {
				fmt.Println("Feature report read failed:", err)
			} else {
				fmt.Println("Feature report: " + hex.EncodeToString(b) + "\n")
			}
		},
	}

	line := liner.NewLiner()
	defer line.Close()
	line.SetCtrlCAborts(true)
	line.SetCompleter(func(line string) (c []string) {
		for cmd, _ := range commands {
			if strings.HasPrefix(cmd, strings.ToLower(line)) {
				c = append(c, cmd+" ")
			}
		}
		return
	})

out:
	for {
		if s, err := line.Prompt("$ "); err == nil {
			if len(s) > 0 {
				line.AppendHistory(s)
				s = strings.ToLower(s)
				for cmd, f := range commands {
					if strings.HasPrefix(s, cmd) {
						s = strings.TrimSpace(s[len(cmd):])
						raw := []byte{}
						if len(s) > 0 {
							raw = make([]byte, len(s)/2, len(s)/2)
							if _, err := hex.Decode(raw, []byte(s)); err != nil {
								fmt.Println("Invalid input:", err)
								fmt.Println(">>", hex.EncodeToString(raw))
								continue out
							}
						}
						f(raw)
						continue out
					}
				}
			}
		} else {
			if err != liner.ErrPromptAborted {
				fmt.Println("Input error:", err)
			}
			break
		}
	}
}

func main() {

	if len(os.Args) == 2 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Println("USAGE:")
		fmt.Printf("  %s                list USB HID devices\n", os.Args[0])
		fmt.Printf("  %s <id> [size]    open USB HID device shell for the given input report size\n", os.Args[0])
		fmt.Printf("  %s -h|--help      show this help\n", os.Args[0])
		fmt.Println()
		return
	}

	// Without arguments - enumerate all HID devices
	if len(os.Args) == 1 {
		found := false
		hid.UsbWalk(func(device hid.Device) {
			info := device.Info()
			fmt.Printf("%04x:%04x:%04x\n", info.Vendor, info.Product, info.Revision)
			found = true
		})
		if !found {
			fmt.Println("No USB HID devices found\n")
		}
		return
	}

	inputReportSize := 1

	if len(os.Args) > 2 {
		if n, err := strconv.Atoi(os.Args[2]); err == nil {
			inputReportSize = n
		} else {
			log.Fatal(err)
		}
	}

	hid.UsbWalk(func(device hid.Device) {
		info := device.Info()
		id := fmt.Sprintf("%04x:%04x:%04x", info.Vendor, info.Product, info.Revision)
		if id != os.Args[1] {
			return
		}

		shell(device, inputReportSize)
	})
}
