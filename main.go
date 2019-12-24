package main // import "lyra.codes/blinken"

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"lyra.codes/blinken/artnet"
	"lyra.codes/blinken/color"
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

	fmt.Println("Polling for nodes")
	if err := transport.Send(artnet.Broadcast, artnet.NewPoll()); err != nil {
		fmt.Printf("Failed to poll: %v\n", err)
		return
	}
	time.Sleep(5 * time.Second)
	if err := transport.Send(artnet.Broadcast, artnet.NewPoll()); err != nil {
		fmt.Printf("Failed to poll: %v\n", err)
		return
	}

	node := <-transport.Nodes()
	fmt.Printf("Polling found %s at (%s; %s)\n", node.ShortName, node.NetworkAddress, node.MAC)

	port := node.Ports[0]
	fmt.Printf("Rendering to port %s (%08b)\n", port.Address, port.Type)

	var q uint8
	s := 0
	uni := make(dmx.Universe, 512)
	colors := make([]color.RGBW, 50)

	for {
		for i := 0; i < len(colors); i++ {
			h := float64((int(float64(i+s) / float64(len(colors)) * 360)) % 360)
			fmt.Printf("%.0f ", h)
			colors[i] = color.HSI{h, 0.6, 0.4}.RGBW()
		}
		fmt.Println()
		fmt.Println(colors[0].String())
		dmx.RGBW(uni).Spread(0, colors)

		if err := transport.Send(node.NetworkAddress, artnet.NewDMX(port.Address, q, uni)); err != nil {
			fmt.Printf("Failed to render: %v\n", err)
			return
		}

		s = (s + 1) % len(colors)
		q++
		time.Sleep(time.Second)
		fmt.Println()
	}
}
