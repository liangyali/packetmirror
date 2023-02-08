package processor

import (
	"fmt"
	"net"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/liangyali/packetmirror/config"
	log "github.com/sirupsen/logrus"
)

type UdpProcessor struct {
}

func (p UdpProcessor) Process(packet gopacket.Packet, config config.Config) {

	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()

	udpLayer := packet.Layer(layers.LayerTypeUDP)
	if udpLayer == nil {
		return
	}

	udpPacket, _ := udpLayer.(*layers.UDP)

	// Dial to the address with UDP
	conn, err := net.DialTimeout("udp", config.OutputUdp, time.Second*1)

	if err != nil {
		log.Error(err)
		return
	}

	log.Debug("forward udp packet to", config.OutputUdp)

	// Send a message to the server
	_, err = conn.Write(udpPacket.Payload)
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Debug("forward sucessed!")
}
