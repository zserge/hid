package hid

const UsbHidClass = 3

type deviceDesc struct {
	Length            uint8
	DescriptorType    uint8
	USB               uint16
	DeviceClass       uint8
	DeviceSubClass    uint8
	DeviceProtocol    uint8
	MaxPacketSize     uint8
	Vendor            uint16
	Product           uint16
	Revision          uint16
	ManufacturerIndex uint8
	ProductIndex      uint8
	SerialIndex       uint8
	NumConfigurations uint8
}

type configDesc struct {
	Length             uint8
	DescriptorType     uint8
	TotalLength        uint16
	NumInterfaces      uint8
	ConfigurationValue uint8
	Configuration      uint8
	Attributes         uint8
	MaxPower           uint8
}

type interfaceDesc struct {
	Length            uint8
	DescriptorType    uint8
	Number            uint8
	AltSetting        uint8
	NumEndpoints      uint8
	InterfaceClass    uint8
	InterfaceSubClass uint8
	InterfaceProtocol uint8
	InterfaceIndex    uint8
}

type endpointDesc struct {
	Length         uint8
	DescriptorType uint8
	Address        uint8
	Attributes     uint8
	MaxPacketSize  uint16
	Interval       uint8
}

type hidReportDesc struct {
	Length         uint8
	DescriptorType uint8
}

const (
	USBDEVFS_CONNECT    = 0x5517
	USBDEVFS_DISCONNECT = 0x5516
	USBDEVFS_CLAIM      = 0x8004550f
	USBDEVFS_RELEASE    = 0x80045510
)

const DevBusUsb = "/dev/bus/usb"

const (
	UsbDescTypeDevice    = 1
	UsbDescTypeConfig    = 2
	UsbDescTypeString    = 3
	UsbDescTypeInterface = 4
	UsbDescTypeEndpoint  = 5
	UsbDescTypeReport    = 33
)
