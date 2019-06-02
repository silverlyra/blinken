package artnet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

// Port is the standard UDP port in the Art-Net specification.
const Port = 6454

// HeaderLength is the length of an Art-Net packet header.
const HeaderLength = 10

// Broadcast is the Art-Net broadcast address.
var Broadcast = &net.UDPAddr{
	IP:   net.IPv4(255, 255, 255, 255),
	Port: Port,
}

// Magic is the prefix of all Art-Net datagrams.
var Magic = []byte("Art-Net\x00")

// Version is an Art-Net protocol version.
type Version uint16

// Version14 is the only known Art-Net protocol version in the wild.
const Version14 Version = 14

func ReadVersion(r io.Reader) (Version, error) {
	var op uint16
	if err := binary.Read(r, binary.BigEndian, op); err != nil {
		return Version(0), err
	}

	return Version(op), nil
}

func ParseVersion(b []byte) Version {
	return Version(binary.BigEndian.Uint16(b))
}

func (o Version) Write(w io.Writer) error {
	op := uint16(o)
	return binary.Write(w, binary.BigEndian, op)
}

// Header is an Art-Net packet header.
type Header struct {
	Operation Operation
}

func (h *Header) Read(r io.Reader) error {
	var magic = make([]byte, len(Magic))
	if _, err := r.Read(magic); err != nil {
		return err
	}

	if !bytes.Equal(Magic, magic) {
		return fmt.Errorf("received invalid magic %q", magic)
	}

	op, err := ReadOperation(r)
	if err != nil {
		return err
	}

	h.Operation = op
	return nil
}

func (h Header) Write(w io.Writer) error {
	if _, err := w.Write(Magic); err != nil {
		return err
	}
	return h.Operation.Write(w)
}

// Operation is an Art-Net op-code.
type Operation uint16

func ReadOperation(r io.Reader) (Operation, error) {
	var op uint16
	if err := binary.Read(r, binary.LittleEndian, &op); err != nil {
		return Operation(0), err
	}

	return Operation(op), nil
}

func ParseOperation(b []byte) Operation {
	return Operation(binary.LittleEndian.Uint16(b))
}

func (o Operation) Write(w io.Writer) error {
	op := uint16(o)
	return binary.Write(w, binary.LittleEndian, op)
}

const (
	OpPoll               Operation = 0x2000
	OpPollReply          Operation = 0x2100
	OpDiagData           Operation = 0x2300
	OpCommand            Operation = 0x2400
	OpDMX                Operation = 0x5000
	OpNZS                Operation = 0x5100
	OpSync               Operation = 0x5200
	OpAddress            Operation = 0x6000
	OpInput              Operation = 0x7000
	OpDeviceTableRequest Operation = 0x8000
	OpDeviceTableData    Operation = 0x8100
	OpDeviceTableControl Operation = 0x8200
	OpRDM                Operation = 0x8300
	OpRDMSub             Operation = 0x8400
)

// Style gives the type of participant in an Art-Net network.
type Style uint8

const (
	StyleNode       Style = 0x00
	StyleController Style = 0x01
	StyleMedia      Style = 0x02
	StyleRoute      Style = 0x03
	StyleBackup     Style = 0x04
	StyleConfig     Style = 0x05
	StyleVisual     Style = 0x06
)

func readString(r io.Reader, capacity int) (string, error) {
	d := make([]byte, capacity)
	if _, err := r.Read(d); err != nil {
		return "", err
	}

	end := bytes.IndexByte(d, 0)
	if end < 0 {
		return "", errors.New("terminating NUL not found")
	}

	return string(d[:end]), nil
}
