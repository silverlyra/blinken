package artnet

import "net"

type Node struct {
	NetworkAddress *net.UDPAddr
	Ports          []NodePort

	ShortName string
	LongName  string

	Style Style
	MAC   net.HardwareAddr
}

type NodePort struct {
	Address Address
	Type    PortType
	Input   PortInput
	Output  PortOutput
}
