package artnet

import (
	"fmt"
)

// Address is an Art-Net v4 DMX universe address.
type Address uint16

const (
	netMask      Address = 0x7F00
	subNetMask   Address = 0x00F0
	universeMask Address = 0x000F
)

func NewAddress(net, subNet, universe uint8) Address {
	return Address(net)<<8 | Address(subNet)<<4 | Address(universe)
}

func (a Address) Net() uint8 {
	return uint8((a & netMask) >> 8)
}

func (a Address) SubNet() uint8 {
	return uint8((a & subNetMask) >> 4)
}

func (a Address) Universe() uint8 {
	return uint8(a & universeMask)
}

func (a Address) String() string {
	return fmt.Sprintf("%d:%d.%d", a.Net(), a.SubNet(), a.Universe())
}
