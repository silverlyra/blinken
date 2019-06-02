package main // import "lyra.codes/blinken"

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lucasb-eyer/go-colorful"

	"lyra.codes/blinken/artnet"
	"lyra.codes/blinken/dmx"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Printf("Caught signal %s\n", sig)
		cancel()

		<-sigs
		os.Exit(2)
	}()

	transport, err := artnet.Listen(ctx, nil)
	if err != nil {
		fmt.Printf("Failed to listen: %v\n", err)
		return
	}

	if err := transport.Send(artnet.Broadcast, artnet.NewPoll()); err != nil {
		fmt.Printf("Failed to poll: %v\n", err)
		return
	}

	node := <-transport.Nodes()
	fmt.Printf("Polling found %s at (%s; %s)\n", node.ShortName, node.NetworkAddress, node.MAC)

	port := node.Ports[0]
	fmt.Printf("Rendering to port %s (%08b)\n", port.Address, port.Type)

	uni := make(dmx.Universe, 512)
	colors := make([]colorful.Color, 10)
	v := 0.3
	d := 0.02
	s := 0
	q := uint8(0)
	for {
		for i := 0; i < 10; i++ {
			h := float64((int(float64(i)/10*360) + s) % 360)
			colors[i] = colorful.Hcl(h, 1.0, v).Clamped()
			fmt.Printf("%.0f ", h)
		}
		fmt.Println()
		dmx.RGB(uni).Spread(0, colors)

		if err := transport.Send(node.NetworkAddress, artnet.NewDMX(port.Address, q, uni)); err != nil {
			fmt.Printf("Failed to render: %v\n", err)
			return
		}
		q++

		v += d
		if v >= 1.0 {
			v = 1.0
			d = -0.02
		} else if v <= 0.3 {
			v = 0.3
			d = 0.02
			s += 2
		}
		fmt.Printf("%d, v=%.1f, d=%.2f\n", q, v, d)

		time.Sleep(25 * time.Millisecond)
	}
}
