package mirror

import (
	"github.com/liangyali/packetmirror/settings"
	"github.com/liangyali/packetmirror/sniffer"
)

type PacketMirror struct {
	device      string //设备名称
	inputFilter string // 流量过滤条件参考BFFilter
	outputHttp  string // http目标地址
	outputUDP   string //udp目标地址
}

type Option func(*PacketMirror)

// WithDevice option
func WithDevice(s string) Option {
	return func(v *PacketMirror) {
		v.device = s
	}
}

// WithInputHttp option
func WithInputFilter(s string) Option {
	return func(v *PacketMirror) {
		v.inputFilter = s
	}
}

// WithOutputHttp option
func WithOutputHttp(s string) Option {
	return func(v *PacketMirror) {
		v.outputHttp = s
	}
}

// WithOutputUdp option
func WithOutputUdp(s string) Option {
	return func(v *PacketMirror) {
		v.outputUDP = s
	}
}

func New(options ...Option) *PacketMirror {
	packetMirror := &PacketMirror{}

	for _, optionFunc := range options {
		optionFunc(packetMirror)
	}

	return packetMirror
}

func (p *PacketMirror) Start() error {

	sniffer := sniffer.New(settings.Settings{
		Device:      p.device,
		InputFilter: p.inputFilter,
		OutputHttp:  p.outputHttp,
		OutputUdp:   p.outputUDP,
	})
	return sniffer.Run()
}
