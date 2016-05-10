package hid

import "time"

//
// General information about the HID device
//
type Info struct {
	Vendor   uint16
	Product  uint16
	Revision uint16

	SubClass uint8
	Protocol uint8

	Interface uint8
}

//
// A common HID device interace
//
type Device interface {
	Open() error
	Close()
	Info() Info
	HIDReport() ([]byte, error)
	SetReport(int, []byte) error
	GetReport(int) ([]byte, error)
	Read(size int, ms time.Duration) ([]byte, error)
	Write(data []byte, ms time.Duration) (int, error)
}
