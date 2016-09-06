package hid

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"
)

type usbDevice struct {
	info Info

	f *os.File

	epIn  int
	epOut int

	inputPacketSize  uint16
	outputPacketSize uint16

	path string
}

func (hid *usbDevice) Open() (err error) {
	if hid.f != nil {
		return errors.New("device is already opened")
	}
	if hid.f, err = os.OpenFile(hid.path, os.O_RDWR, 0644); err != nil {
		return
	} else {
		return hid.claim()
	}
}

func (hid *usbDevice) Close() {
	if hid.f != nil {
		hid.release()
		hid.f.Close()
		hid.f = nil
	}
}

func (hid *usbDevice) Info() Info {
	return hid.info
}

func (hid *usbDevice) ioctl(n uint32, arg interface{}) (int, error) {
	b := new(bytes.Buffer)
	if err := binary.Write(b, binary.LittleEndian, arg); err != nil {
		return -1, err
	}
	r, _, err := syscall.Syscall6(syscall.SYS_IOCTL,
		uintptr(hid.f.Fd()), uintptr(n),
		uintptr(unsafe.Pointer(&(b.Bytes()[0]))), 0, 0, 0)
	return int(r), err
}

func (hid *usbDevice) claim() error {
	ifno := uint32(hid.info.Interface)
	if r, errno := hid.ioctl(USBDEVFS_IOCTL, &usbfsIoctl{
		Interface: ifno,
		IoctlCode: USBDEVFS_DISCONNECT,
		Data:      0,
	}); r == -1 {
		log.Println("driver disconnect failed:", r, errno)
	}

	if r, errno := hid.ioctl(USBDEVFS_CLAIM, &ifno); r == -1 {
		return errno
	} else {
		return nil
	}
	return nil
}

func (hid *usbDevice) release() error {
	ifno := uint32(hid.info.Interface)
	if r, errno := hid.ioctl(USBDEVFS_RELEASE, &ifno); r == -1 {
		return errno
	}

	if r, errno := hid.ioctl(USBDEVFS_IOCTL, &usbfsIoctl{
		Interface: ifno,
		IoctlCode: USBDEVFS_CONNECT,
		Data:      0,
	}); r == -1 {
		log.Println("driver connect failed:", r, errno)
	}
	return nil
}

func (hid *usbDevice) ctrl(rtype, req, val, index int, data []byte, t int) (int, error) {
	s := usbfsCtrl{
		ReqType: uint8(rtype),
		Req:     uint8(req),
		Value:   uint16(val),
		Index:   uint16(index),
		Len:     uint16(len(data)),
		Timeout: uint32(t),
		Data:    slicePtr(data),
	}
	if r, err := hid.ioctl(USBDEVFS_CONTROL, &s); r == -1 {
		return -1, err
	} else {
		return r, nil
	}
}

func (hid *usbDevice) intr(ep int, data []byte, t int) (int, error) {
	if r, err := hid.ioctl(USBDEVFS_BULK, &usbfsBulk{
		Endpoint: uint32(ep),
		Len:      uint32(len(data)),
		Timeout:  uint32(t),
		Data:     slicePtr(data),
	}); r == -1 {
		return -1, err
	} else {
		return r, nil
	}
}

func (hid *usbDevice) Read(size int, timeout time.Duration) ([]byte, error) {
	if size < 0 {
		size = int(hid.inputPacketSize)
	}
	data := make([]byte, size, size)
	ms := timeout / (1 * time.Millisecond)
	n, err := hid.intr(hid.epIn, data, int(ms))
	if err == nil {
		return data[:n], nil
	} else {
		return nil, err
	}
}

func (hid *usbDevice) Write(data []byte, timeout time.Duration) (int, error) {
	if hid.epOut > 0 {
		ms := timeout / (1 * time.Millisecond)
		return hid.intr(hid.epOut, data, int(ms))
	} else {
		return hid.ctrl(0x21, 0x09, 2<<8+0, int(hid.info.Interface), data, len(data))
	}
}

func (hid *usbDevice) HIDReport() ([]byte, error) {
	buf := make([]byte, 256, 256)
	// In transfer, recepient interface, GetDescriptor, HidReport type
	n, err := hid.ctrl(0x81, 0x06, 0x22<<8+int(hid.info.Interface), 0, buf, 1000)
	if err != nil {
		return nil, err
	} else {
		return buf[:n], nil
	}
}

func (hid *usbDevice) GetReport(report int) ([]byte, error) {
	buf := make([]byte, 256, 256)
	// 10100001, GET_REPORT, type*256+id, intf, len, data
	n, err := hid.ctrl(0xa1, 0x01, 3<<8+report, int(hid.info.Interface), buf, 1000)
	if err != nil {
		return nil, err
	} else {
		return buf[:n], nil
	}
}

func (hid *usbDevice) SetReport(report int, data []byte) error {
	// 00100001, SET_REPORT, type*256+id, intf, len, data
	_, err := hid.ctrl(0x21, 0x09, 3<<8+report, int(hid.info.Interface), data, 1000)
	return err
}

//
// Enumeration
//

func cast(b []byte, to interface{}) error {
	r := bytes.NewBuffer(b)
	return binary.Read(r, binary.LittleEndian, to)
}

func walker(path string, cb func(Device)) error {
	if desc, err := ioutil.ReadFile(path); err != nil {
		return err
	} else {
		r := bytes.NewBuffer(desc)
		expected := map[byte]bool{
			UsbDescTypeDevice: true,
		}
		devDesc := deviceDesc{}
		var device *usbDevice
		for r.Len() > 0 {
			if length, err := r.ReadByte(); err != nil {
				return err
			} else if err := r.UnreadByte(); err != nil {
				return err
			} else {
				body := make([]byte, length, length)
				if n, err := r.Read(body); err != nil {
					return err
				} else if n != int(length) || length < 2 {
					return errors.New("short read")
				} else {
					if !expected[body[1]] {
						continue
					}
					switch body[1] {
					case UsbDescTypeDevice:
						expected[UsbDescTypeDevice] = false
						expected[UsbDescTypeConfig] = true
						if err := cast(body, &devDesc); err != nil {
							return err
						}
						//info := Info{
						//}
					case UsbDescTypeConfig:
						expected[UsbDescTypeInterface] = true
						expected[UsbDescTypeReport] = false
						expected[UsbDescTypeEndpoint] = false
						// Device left from the previous config
						if device != nil {
							cb(device)
							device = nil
						}
					case UsbDescTypeInterface:
						if device != nil {
							cb(device)
							device = nil
						}
						expected[UsbDescTypeEndpoint] = true
						expected[UsbDescTypeReport] = true
						i := &interfaceDesc{}
						if err := cast(body, i); err != nil {
							return err
						}
						if i.InterfaceClass == UsbHidClass {
							device = &usbDevice{
								info: Info{
									Vendor:    devDesc.Vendor,
									Product:   devDesc.Product,
									Revision:  devDesc.Revision,
									SubClass:  i.InterfaceSubClass,
									Protocol:  i.InterfaceProtocol,
									Interface: i.Number,
								},
								path: path,
							}
						}
					case UsbDescTypeEndpoint:
						if device != nil {
							if device.epIn != 0 && device.epOut != 0 {
								cb(device)
								device.epIn = 0
								device.epOut = 0
							}
							e := &endpointDesc{}
							if err := cast(body, e); err != nil {
								return err
							}
							if e.Address > 0x80 && device.epIn == 0 {
								device.epIn = int(e.Address)
								device.inputPacketSize = e.MaxPacketSize
							} else if e.Address < 0x80 && device.epOut == 0 {
								device.epOut = int(e.Address)
								device.outputPacketSize = e.MaxPacketSize
							}
						}
					}
				}
			}
		}
		if device != nil {
			cb(device)
		}
	}
	return nil
}

func UsbWalk(cb func(Device)) {
	filepath.Walk(DevBusUsb, func(f string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() {
			return nil
		}
		if err := walker(f, cb); err != nil {
			log.Println("UsbWalk: ", err)
		}
		return nil
	})
}
