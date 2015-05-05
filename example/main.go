package main

import (
	"log"

	"github.com/zserge/hid"
)

func main() {
	hid.UsbWalk(func(device hid.Device) {
		log.Printf("%+v\n", device.Info())
		if err := device.Open(); err != nil {
			log.Println("Open error: ", err)
			return
		}
		defer device.Close()

		log.Println(device.HIDReport())

		for i := 0; i < 10; i++ {
			log.Println(device.Read(8, 0))
		}
	})
}
