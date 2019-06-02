package artnet

import (
	"io"

	"lyra.codes/blinken/artnet/wire"
)

type Packet interface {
	Read(r wire.Reader) error
	Write(w io.Writer) error
}
