package processor

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/liangyali/packetmirror/settings"
)

func init() {
	Register("http", &HttpProcessor{})
}

type HttpProcessor struct {
}

func (p HttpProcessor) Process(packet gopacket.Packet, settings settings.Settings) {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(time.Now().Format("2006-01-02 15:04:05.000"), err)
		}
	}()

	applicationLayer := packet.ApplicationLayer()

	if applicationLayer == nil {
		return
	}

	payload := string(applicationLayer.Payload())

	reader := bufio.NewReader(strings.NewReader(payload))
	request, err := http.ReadRequest(reader)

	if err != nil {
		return
	}

	URL, _ := url.Parse(settings.OutputHttp)

	request.URL.Host = URL.Host
	request.URL.Scheme = URL.Scheme
	request.RequestURI = ""

	delete(request.Header, "Accept-Encoding")
	delete(request.Header, "Content-Length")

	// set X-Forwarded-For
	request.Header.Set("X-Forwarded-For", request.RemoteAddr)

	client := &http.Client{}
	resp, err := client.Do(request.WithContext(context.Background()))
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
}
