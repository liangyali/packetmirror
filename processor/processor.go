package processor

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
	"github.com/liangyali/packetmirror/config"
	log "github.com/sirupsen/logrus"
)

var (
	rwLock     sync.RWMutex
	processors = make(map[string]PacketProcessor)
)

type PacketProcessor interface {
	Process(packet gopacket.Packet, settings config.Config)
}

// httpStreamFactory implements tcpassembly.StreamFactory
type httpStreamFactory struct {
	config config.Config
}

// httpStream will handle the actual decoding of http requests.
type httpStream struct {
	net, transport gopacket.Flow
	r              tcpreader.ReaderStream
}

func (h *httpStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	hstream := &httpStream{
		net:       net,
		transport: transport,
		r:         tcpreader.NewReaderStream(),
	}
	go hstream.run(h.config)

	// ReaderStream implements tcpassembly.Stream, so we can return a pointer to it.
	return &hstream.r
}

func (h *httpStream) run(config config.Config) {
	buf := bufio.NewReader(&h.r)
	for {
		req, err := http.ReadRequest(buf)
		if err == io.EOF {
			// We must read until we see an EOF... very important!
			// return
			return
		} else if err != nil {
			// pass
		} else {
			h.request(req.Clone(context.Background()), config)
		}
	}
}

func (h *httpStream) request(req *http.Request, config config.Config) {

	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()

	if config.Debug {
		var builder strings.Builder

		builder.WriteString("TCP ")
		builder.WriteString("【")
		builder.WriteString(h.net.Src().String())
		builder.WriteString(":")
		builder.WriteString(h.transport.Src().String())
		builder.WriteString("】")
		builder.WriteString("=>")

		builder.WriteString("【")
		builder.WriteString(h.net.Dst().String())
		builder.WriteString(":")
		builder.WriteString(h.transport.Dst().String())
		builder.WriteString("】")

		log.WithFields(log.Fields{
			"url":    req.RequestURI,
			"method": req.Method,
		}).Debug(builder.String())
	}

	URL, err := url.Parse(config.OutputHttp)
	if err != nil {
		panic(err)
	}

	req.URL.Host = URL.Host
	req.URL.Scheme = URL.Scheme
	req.RequestURI = ""

	req.URL.Host = URL.Host
	req.URL.Scheme = URL.Scheme
	req.RequestURI = ""

	delete(req.Header, "Accept-Encoding")
	delete(req.Header, "Content-Length")

	req.Header.Set("Packet-Mirror-Ts", strconv.Itoa(int(time.Now().Unix())))

	tr := &http.Transport{
		MaxIdleConns:          10,
		IdleConnTimeout:       30 * time.Second,
		DisableCompression:    true,
		ResponseHeaderTimeout: 5 * time.Second,
	}
	client := &http.Client{Transport: tr, Timeout: 3 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
		panic(err)
	}

	defer resp.Body.Close()
}

var udpProcessor = UdpProcessor{}

func Process(packets chan gopacket.Packet, config config.Config) {

	// Set up assembly
	streamFactory := &httpStreamFactory{config: config}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)

	ticker := time.Tick(time.Minute)
	for {
		select {
		case packet := <-packets:
			// A nil packet indicates the end of a pcap file.
			if packet == nil {
				return
			}

			if packet.NetworkLayer() == nil || packet.TransportLayer() == nil || packet.TransportLayer().LayerType() != layers.LayerTypeTCP {

				udpLayer := packet.Layer(layers.LayerTypeUDP)
				if udpLayer != nil {
					udpProcessor.Process(packet, config)
					continue
				}

				log.Println("Unusable packet")
				continue
			}

			tcp := packet.TransportLayer().(*layers.TCP)
			assembler.AssembleWithTimestamp(packet.NetworkLayer().NetworkFlow(), tcp, packet.Metadata().Timestamp)

		case <-ticker:
			// Every minute, flush connections that haven't seen activity in the past 2 minutes.
			assembler.FlushOlderThan(time.Now().Add(time.Minute * -2))
		}
	}
}
