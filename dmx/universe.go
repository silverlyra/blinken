// Package dmx implements DMX data types.
package dmx

// Channel is a single DMX channel value.
type Channel = uint8

// Universe is a DMX universe of up to 512 channels.
type Universe []Channel
