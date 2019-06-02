package wire

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
)

type FieldError struct {
	Field string
	Err   error
}

func (e FieldError) Error() string {
	return fmt.Sprintf("field %q: %v", e.Field, e.Err)
}

func Parse(r Reader) *Parser {
	return &Parser{Reader: r}
}

type Reader interface {
	io.Reader
	io.ByteReader

	Next(n int) []byte
}

type Parser struct {
	Reader Reader
	err    error
}

func (p Parser) Err() error {
	return p.err
}

func (p *Parser) error(name string, err error) {
	if p.err != nil {
		return
	}

	p.err = &FieldError{Field: name, Err: err}
}

func (p *Parser) Int8(name string) uint8 {
	b, err := p.Reader.ReadByte()
	if err != nil {
		p.error(name, err)
		return 0
	}

	return uint8(b)
}

func (p *Parser) Int16(name string, ord binary.ByteOrder) uint16 {
	b := []byte{0, 0}
	_, err := p.Reader.Read(b)
	if err != nil {
		p.error(name, err)
		return 0
	}

	return ord.Uint16(b)
}

func (p *Parser) Int32(name string, ord binary.ByteOrder) uint32 {
	b := []byte{0, 0, 0, 0}
	_, err := p.Reader.Read(b)
	if err != nil {
		p.error(name, err)
		return 0
	}

	return ord.Uint32(b)
}

func (p *Parser) String(name string, capacity int) string {
	d := make([]byte, capacity)
	if _, err := p.Reader.Read(d); err != nil {
		p.error(name, err)
		return ""
	}

	end := bytes.IndexByte(d, 0)
	if end < 0 {
		p.error(name, errors.New("terminating NUL not found"))
	}

	return string(d[:end])
}

func (p *Parser) IPv4(name string) net.IP {
	b := []byte{0, 0, 0, 0}
	_, err := p.Reader.Read(b)
	if err != nil {
		p.error(name, err)
		return nil
	}

	return net.IP(b)
}

func (p *Parser) MAC(name string) net.HardwareAddr {
	b := []byte{0, 0, 0, 0, 0, 0}
	_, err := p.Reader.Read(b)
	if err != nil {
		p.error(name, err)
		return nil
	}

	return net.HardwareAddr(b)
}

func (p *Parser) Skip(name string, n int) {
	p.Reader.Next(n)
}
