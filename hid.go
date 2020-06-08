package hid

import (
	"io/ioutil"
	"log"
	"time"
)

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
	Bus       int
	Device    int
}

type getReportOptions struct {
	bufferSize uint16
}

// GetReportOption is an optional argument to the GetReport function.
type GetReportOption func(*getReportOptions)

// WithBufferSize returns an option for the GetReport function that sets the
// buffer size to use. The buffer size defaults to 256 bytes, which is not
// suitable for some devices. Using this optional argument, the caller can use
// a smaller or larger buffer.
func WithBufferSize(sz uint16) GetReportOption {
	return func(opts *getReportOptions) {
		opts.bufferSize = sz
	}
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
	GetReport(int, ...GetReportOption) ([]byte, error)
	Read(size int, ms time.Duration) ([]byte, error)
	Write(data []byte, ms time.Duration) (int, error)
	Ctrl(rtype, req, val, index int, data []byte, t int) (int, error)
}

// Default Logger setting
var Logger = log.New(ioutil.Discard, "hid", log.LstdFlags)
