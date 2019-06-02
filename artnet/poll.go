package artnet

import (
	"encoding/binary"
	"errors"
	"io"
	"net"

	"lyra.codes/blinken/artnet/wire"
)

var (
	pollHeader      = Header{Operation: OpPoll}
	pollReplyHeader = Header{Operation: OpPollReply}
)

// Poll is the contents of an OpPoll message.
type Poll struct {
	Header
	Version  Version
	TalkToMe TalkToMe
	Priority DiagnosticPriority
}

// NewPoll creates a new Poll operation.
func NewPoll(options ...PollOption) *Poll {
	p := &Poll{
		Header:  pollHeader,
		Version: Version14,
	}

	for _, opt := range options {
		opt(p)
	}

	return p
}

type PollOption func(p *Poll)

// PollPush requests nodes to send additional poll replies when the node changes.
func PollPush() PollOption {
	return func(p *Poll) {
		TalkToMePushChanges.Set(&p.TalkToMe)
	}
}

// PollDiagnostics requests diagnostic messages from nodes.
func PollDiagnostics(priority DiagnosticPriority, broadcast bool) PollOption {
	return func(p *Poll) {
		TalkToMeSendDiagnostics.Set(&p.TalkToMe)
		if broadcast {
			TalkToMeBroadcastDiagnostics.Set(&p.TalkToMe)
		}
		p.Priority = priority
	}
}

func (p *Poll) Read(r wire.Reader) error {
	p.Header.Read(r)

	parser := wire.Parse(r)
	p.Version = Version(parser.Int16("Version", binary.BigEndian))
	p.TalkToMe = TalkToMe(parser.Int8("TalkToMe"))
	p.Priority = DiagnosticPriority(parser.Int8("Priority"))

	return parser.Err()
}

func (p *Poll) Write(w io.Writer) error {
	p.Header.Write(w)

	return wire.Build(w).
		Int16("Version", uint16(p.Version), binary.BigEndian).
		Int8("TalkToMe", uint8(p.TalkToMe)).
		Int8("Priority", uint8(p.Priority)).
		Err()
}

// TalkToMe is the Node behavior options that can be sent in a Poll operation.
type TalkToMe uint8

const (
	TalkToMeVLC                  TalkToMe = 0x8
	TalkToMeBroadcastDiagnostics TalkToMe = 0x4
	TalkToMeSendDiagnostics      TalkToMe = 0x2
	TalkToMePushChanges          TalkToMe = 0x1
)

func (f TalkToMe) Enabled(v TalkToMe) bool {
	return (v & f) != 0
}

func (f TalkToMe) Add(v TalkToMe) TalkToMe {
	return v | f
}

func (f TalkToMe) Set(v *TalkToMe) {
	*v |= f
}

// DiagnosticPriority is the priority of a diagnostic message from a node.
type DiagnosticPriority uint8

const (
	DPLow      DiagnosticPriority = 0x10
	DPMed      DiagnosticPriority = 0x40
	DPHigh     DiagnosticPriority = 0x80
	DPCritical DiagnosticPriority = 0xE0
	DPVolatile DiagnosticPriority = 0xF0
	DPAll      DiagnosticPriority = 0
)

// PollReply is a reponse to a Poll message from an Art-Net device.
type PollReply struct {
	Header
	Node            net.UDPAddr
	FirmwareVersion uint16

	NetSwitch uint8
	SubSwitch uint8

	OEM         uint16
	UBEAVersion uint8

	Status1          uint8
	ESTAManufacturer uint16

	ShortName  string
	LongName   string
	NodeReport string

	PortCount       uint16
	PortTypes       []PortType
	PortInputs      []PortInput
	PortOutputs     []PortOutput
	InputUniverses  []uint8
	OutputUniverses []uint8

	Video  uint8
	Macro  uint8
	Remote uint8

	Style     Style
	MAC       net.HardwareAddr
	BindIP    net.IP
	BindIndex uint8

	Status2 uint8
}

type PortType uint8

type PortInput uint8

type PortOutput uint8

func (p *PollReply) ToNode() *Node {
	return &Node{
		NetworkAddress: &p.Node,
		Ports:          p.Ports(),
		ShortName:      p.ShortName,
		LongName:       p.LongName,
		Style:          p.Style,
		MAC:            p.MAC,
	}
}

func (p *PollReply) Ports() []NodePort {
	count := int(p.PortCount)
	ports := make([]NodePort, 0, count)

	for i := 0; i < count; i++ {
		// TODO(lyra): can the input and output universes really be different?
		ports = append(ports, NodePort{
			Address: NewAddress(p.NetSwitch, p.SubSwitch, p.OutputUniverses[i]),
			Type:    p.PortTypes[i],
			Input:   p.PortInputs[i],
			Output:  p.PortOutputs[i],
		})
	}

	return ports
}

func (p *PollReply) Read(r wire.Reader) error {
	p.Header.Read(r)

	parser := wire.Parse(r)
	p.Node = net.UDPAddr{
		IP:   parser.IPv4("Node.IP"),
		Port: int(parser.Int16("Node.Port", binary.LittleEndian)),
	}
	p.FirmwareVersion = parser.Int16("FirmwareVersion", binary.BigEndian)

	p.NetSwitch = parser.Int8("NetSwitch")
	p.SubSwitch = parser.Int8("SubSwitch")

	p.OEM = parser.Int16("OEM", binary.BigEndian)
	p.UBEAVersion = parser.Int8("UBEAVersion")

	p.Status1 = parser.Int8("Status1")
	p.ESTAManufacturer = parser.Int16("ESTAManufacturer", binary.BigEndian)

	p.ShortName = parser.String("ShortName", 18)
	p.LongName = parser.String("LongName", 64)
	p.NodeReport = parser.String("NodeReport", 64)

	p.PortCount = parser.Int16("PortCount", binary.BigEndian)
	p.PortTypes = make([]PortType, 4)
	for i := 0; i < 4; i++ {
		p.PortTypes[i] = PortType(parser.Int8("PortTypes"))
	}
	p.PortInputs = make([]PortInput, 4)
	for i := 0; i < 4; i++ {
		p.PortInputs[i] = PortInput(parser.Int8("PortInputs"))
	}
	p.PortOutputs = make([]PortOutput, 4)
	for i := 0; i < 4; i++ {
		p.PortOutputs[i] = PortOutput(parser.Int8("PortOutputs"))
	}
	p.InputUniverses = make([]uint8, 4)
	for i := 0; i < 4; i++ {
		p.InputUniverses[i] = parser.Int8("InputUniverses")
	}
	p.OutputUniverses = make([]uint8, 4)
	for i := 0; i < 4; i++ {
		p.OutputUniverses[i] = parser.Int8("OutputUniverses")
	}

	p.Video = parser.Int8("Video")
	p.Macro = parser.Int8("Macro")
	p.Remote = parser.Int8("Remote")

	parser.Skip("Spare", 3)
	p.Style = Style(parser.Int8("Style"))
	p.MAC = parser.MAC("MAC")
	p.BindIP = parser.IPv4("BindIP")
	p.BindIndex = parser.Int8("BindIndex")

	p.Status2 = parser.Int8("Status2")

	return parser.Err()
}

func (p *PollReply) Write(w io.Writer) error {
	return errors.New("PollReply.Write() not implemented")
}
