package artnet

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"sync"
)

type Transport interface {
	Send(to *net.UDPAddr, packet Packet) error
	Nodes() <-chan *Node
}

func Listen(ctx context.Context, addr *net.UDPAddr) (Transport, error) {
	if addr == nil {
		addr = &net.UDPAddr{Port: Port}
	}

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return nil, err
	}

	t := &networkTransport{
		ctx:  ctx,
		conn: conn,

		pool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, 2048)
			},
		},
		recv:  make(chan networkMessage),
		nodes: make(chan *Node, 1),
	}

	go t.receive()
	go t.process()

	return t, nil
}

type networkTransport struct {
	ctx  context.Context
	conn *net.UDPConn

	pool  *sync.Pool
	recv  chan networkMessage
	nodes chan *Node
}

func (t *networkTransport) Send(to *net.UDPAddr, packet Packet) error {
	buf := bytes.Buffer{}
	if err := packet.Write(&buf); err != nil {
		return err
	}

	if _, err := t.conn.WriteToUDP(buf.Bytes(), to); err != nil {
		return err
	}

	return nil
}

func (t *networkTransport) Nodes() <-chan *Node {
	return t.nodes
}

type networkMessage struct {
	addr *net.UDPAddr
	body []byte
}

func (t *networkTransport) buffer() ([]byte, func()) {
	buf := t.pool.Get().([]byte)
	put := func() {
		t.pool.Put(buf)
	}

	return buf, put
}

func (t *networkTransport) process() {
	done := t.ctx.Done()

	for {
		select {
		case msg, ok := <-t.recv:
			if !ok {
				fmt.Println("Socket closed")
				return
			}

			t.handle(msg.addr, msg.body)
		case <-done:
			fmt.Println("Shutting down")
			t.conn.Close() // TODO(lyra): blackout on shutdown
			return
		}
	}
}

func (t *networkTransport) handle(from *net.UDPAddr, body []byte) {
	head := Header{}
	hbuf := bytes.NewBuffer(body[:HeaderLength])
	if err := head.Read(hbuf); err != nil {
		fmt.Printf("invalid packet: %v\n", err)
		return
	}

	fmt.Printf("Got 0x%04x message from %s\n", head.Operation, from)
	fmt.Println(hex.Dump(body))

	buf := bytes.NewBuffer(body)
	switch head.Operation {
	case OpPollReply:
		p := &PollReply{Header: head}
		if err := p.Read(buf); err != nil {
			fmt.Printf("Error reading OpPollReply: %v\n", err)
			return
		}

		t.nodes <- p.ToNode()
	default:
		fmt.Println("unknown operation")
	}
}

func (t *networkTransport) receive() {
	defer close(t.recv)
	defer close(t.nodes)

	for {
		err := t.receiveNext()
		if err != nil {
			if nerr := err.(net.Error); nerr != nil && nerr.Temporary() {
				continue
			}

			return
		}
	}
}

func (t *networkTransport) receiveNext() error {
	buf, release := t.buffer()
	defer release()

	n, from, err := t.conn.ReadFromUDP(buf)
	if n > 0 {
		t.recv <- networkMessage{from, buf[:n]}
	}

	return err
}
