package sniffer

import (
	"context"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/liangyali/packetmirror/config"
	"github.com/liangyali/packetmirror/processor"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Sniffer struct {
	config config.Config
}

func New(config config.Config) *Sniffer {
	log.Debug(config)
	return &Sniffer{
		config: config,
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
		log.Info("started")
		return s.sniffStatic(ctx)
	})

	return g.Wait()
}

// sniffStatic performs the sniffing work on a single static interface.
func (s *Sniffer) sniffStatic(ctx context.Context) error {
	handle, err := pcap.OpenLive(s.config.Device, 1024*5, true, 2*time.Second)
	if err != nil {
		log.Error(err)
		return err
	}

	err = handle.SetBPFFilter(s.config.InputFilter)
	if err != nil {
		log.Error(err)
		handle.Close()
		return err
	}
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	processor.Process(packetSource.Packets(), s.config)
	return nil
}
