package artnet

import (
	"encoding/binary"
	"io"

	"lyra.codes/blinken/artnet/wire"
	"lyra.codes/blinken/dmx"
)

var (
	dmxHeader = Header{Operation: OpDMX}
)

// DMX is the contents of an OpDMX message.
type DMX struct {
	Header
	Version  Version
	Sequence uint8
	Input    uint8
	Address  Address
	Length   uint16
	Data     dmx.Universe
}

// NewDMX creates a new DMX operation.
func NewDMX(dest Address, seq uint8, channels dmx.Universe) *DMX {
	return &DMX{
		Header:   dmxHeader,
		Version:  Version14,
		Sequence: seq,
		Input:    0,
		Address:  dest,
		Length:   uint16(len(channels)),
		Data:     channels,
	}
}

func (p *DMX) Read(r wire.Reader) error {
	p.Header.Read(r)

	parser := wire.Parse(r)
	p.Version = Version(parser.Int16("Version", binary.BigEndian))
	p.Sequence = parser.Int8("Sequence")
	p.Input = parser.Int8("Sequence")
	p.Address = Address(parser.Int16("Address", binary.LittleEndian))
	p.Length = parser.Int16("Length", binary.BigEndian)
	if parser.Err() != nil {
		return parser.Err()
	}

	p.Data = make(dmx.Universe, int(p.Length))
	_, err := r.Read([]byte(p.Data))
	return err
}

func (p *DMX) Write(w io.Writer) error {
	p.Header.Write(w)

	err := wire.Build(w).
		Int16("Version", uint16(p.Version), binary.BigEndian).
		Int8("Sequence", uint8(p.Sequence)).
		Int8("Input", uint8(p.Input)).
		Int16("Address", uint16(p.Address), binary.LittleEndian).
		Int16("Length", p.Length, binary.BigEndian).
		Err()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(p.Data))
	return err
}
