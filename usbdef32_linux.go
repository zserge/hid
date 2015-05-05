// +build !amd64

package hid

import (
	"unsafe"
)

const (
	USBDEVFS_IOCTL   = 0xc00c5512
	USBDEVFS_BULK    = 0xc0105502
	USBDEVFS_CONTROL = 0xc0105500
)

type usbfsIoctl struct {
	Interface uint32
	IoctlCode uint32
	Data      uint32
}

type usbfsCtrl struct {
	ReqType uint8
	Req     uint8
	Value   uint16
	Index   uint16
	Len     uint16
	Timeout uint32
	Data    uint32
}

type usbfsBulk struct {
	Endpoint uint32
	Len      uint32
	Timeout  uint32
	Data     uint32
}

func slicePtr(b []byte) uint32 {
	return uint32(uintptr(unsafe.Pointer(&b[0])))
}
