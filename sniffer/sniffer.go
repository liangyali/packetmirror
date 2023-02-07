package sniffer

import (
	"context"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/liangyali/packetmirror/processor"
	"github.com/liangyali/packetmirror/settings"
	"golang.org/x/sync/errgroup"
)

type Sniffer struct {
	settings settings.Settings
}

func New(settings settings.Settings) *Sniffer {
	return &Sniffer{
		settings: settings,
	}
}

func validatePcapFilter(expr string) error {
	if expr == "" {
		return nil
	}

	_, err := pcap.NewBPF(layers.LinkTypeEthernet, 65535, expr)
	return err
}

func (s *Sniffer) Run() error {

	g, ctx := errgroup.WithContext(context.Background())
	g.Go(func() error {
		return s.sniffStatic(ctx)
	})

	return g.Wait()
}

// sniffStatic performs the sniffing work on a single static interface.
func (s *Sniffer) sniffStatic(ctx context.Context) error {
	handle, err := pcap.OpenLive(s.settings.Device, 1024, true, 500*time.Millisecond)
	if err != nil {
		return err
	}

	err = handle.SetBPFFilter(s.settings.InputFilter)
	if err != nil {
		handle.Close()
		return err
	}
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		go func(packet gopacket.Packet, options settings.Settings) {
			processor.Process(packet, options)
		}(packet, s.settings)
	}

	return nil
}
