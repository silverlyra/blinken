package wire

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

type Builder struct {
	Writer io.Writer
	err    error
}

func Build(w io.Writer) *Builder {
	return &Builder{Writer: w}
}

func (b Builder) Err() error {
	return b.err
}

func (b *Builder) error(name string, err error) {
	if b.err != nil {
		return
	}

	b.err = &FieldError{Field: name, Err: err}
}

func (b *Builder) Int8(name string, v uint8) *Builder {
	if _, err := b.Writer.Write([]byte{v}); err != nil {
		b.error(name, err)
	}
	return b
}

func (b *Builder) Int16(name string, v uint16, ord binary.ByteOrder) *Builder {
	d := []byte{0, 0}
	ord.PutUint16(d, v)
	if _, err := b.Writer.Write(d); err != nil {
		b.error(name, err)
	}
	return b
}

func (b *Builder) Int32(name string, v uint32, ord binary.ByteOrder) *Builder {
	d := []byte{0, 0, 0, 0}
	ord.PutUint32(d, v)
	if _, err := b.Writer.Write(d); err != nil {
		b.error(name, err)
	}
	return b
}

func (b *Builder) String(name, v string, capacity int) *Builder {
	if len(v)+1 > capacity {
		b.error(name, fmt.Errorf("string is longer than capacity %d", capacity))
		return b
	}

	d := make([]byte, capacity)
	copy(d, []byte(v))
	if _, err := b.Writer.Write(d); err != nil {
		b.error(name, err)
	}

	return b
}

func (b *Builder) IPv4(name string, v net.IP) *Builder {
	ip := v.To4()
	if ip != nil {
		b.error(name, fmt.Errorf("IP %q is not an IPv4 address", v))
		return b
	}

	if _, err := b.Writer.Write([]byte(ip)); err != nil {
		b.error(name, err)
	}

	return b
}

func (b *Builder) MAC(name string, v net.HardwareAddr) *Builder {
	if _, err := b.Writer.Write([]byte(v)); err != nil {
		b.error(name, err)
	}

	return b
}
