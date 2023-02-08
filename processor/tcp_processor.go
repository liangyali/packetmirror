package processor

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/liangyali/packetmirror/config"
	log "github.com/sirupsen/logrus"
)

type TcpProcessor struct {
}

func (p TcpProcessor) Process(packet gopacket.Packet, settings config.Config) {

	// defer func() {
	// 	if err := recover(); err != nil {
	// 		log.Error(err)
	// 	}
	// }()

	if packet.Layer(layers.LayerTypeTCP) == nil {
		return
	}

	applicationLayer := packet.ApplicationLayer()

	if applicationLayer == nil {
		log.Debug("applicationLayer is nil")
		return
	}

	fmt.Println(string(applicationLayer.Payload()))

	reader := bufio.NewReader(bytes.NewReader(applicationLayer.Payload()))
	request, err := http.ReadRequest(reader)

	if err != nil {
		log.Warn("cannot parse request ignore packet")
		return
	}

	log.WithFields(log.Fields{
		"method":     request.Method,
		"requestURI": request.RequestURI,
	}).Info("process http request")

	URL, _ := url.Parse(settings.OutputHttp)

	request.URL.Host = URL.Host
	request.URL.Scheme = URL.Scheme
	request.RequestURI = ""

	delete(request.Header, "Accept-Encoding")
	delete(request.Header, "Content-Length")

	log.Debug("forward http packet to:", settings.OutputHttp)
	client := &http.Client{}

	resp, err := client.Do(request)
	if err != nil {
		log.Error(err)
		panic(err)
	}

	log.WithFields(log.Fields{
		"dst": settings.OutputHttp,
	}).Debug("forward http packet success!")
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(body)

}
