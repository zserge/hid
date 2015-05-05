// +build amd64

package hid

import (
	"unsafe"
)

const (
	USBDEVFS_IOCTL   = 0xc0105512
	USBDEVFS_BULK    = 0xc0185502
	USBDEVFS_CONTROL = 0xc0185500
)

type usbfsIoctl struct {
	Interface uint32
	IoctlCode uint32
	Data      uint64
}

type usbfsCtrl struct {
	ReqType uint8
	Req     uint8
	Value   uint16
	Index   uint16
	Len     uint16
	Timeout uint32
	_       uint32
	Data    uint64 // FIXME
}

type usbfsBulk struct {
	Endpoint uint32
	Len      uint32
	Timeout  uint32
	_        uint32
	Data     uint64
}

func slicePtr(b []byte) uint64 {
	return uint64(uintptr(unsafe.Pointer(&b[0])))
}
