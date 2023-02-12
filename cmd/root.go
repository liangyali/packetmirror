package cmd

import (
	"flag"

	"github.com/liangyali/packetmirror/internal/mirror"
	log "github.com/sirupsen/logrus"
)

func Run() error {

	device := flag.String("device", "", "device name for interface")
	debug := flag.Bool("debug", false, "debug mode default(false)")
	inputFilter := flag.String("input-filter", "", "bbfilter query")
	outputHttp := flag.String("output-http", "http://127.0.0.1:80", "http forward address (http://127.0.0.1:80).")
	outputUdp := flag.String("output-udp", "127.0.0.1:80", "udf forward addess (127.0.0.1:80).")

	flag.Parse()

	// 参数初始化
	var options []mirror.Option

	// 设置网卡名称
	if *device != "" {
		options = append(options, mirror.WithDevice(*device))
	}

	// 设置流量过滤器
	if *inputFilter != "" {
		options = append(options, mirror.WithInputFilter(*inputFilter))
	}

	// 设置http转发的目标地址
	if *outputHttp != "" {
		options = append(options, mirror.WithOutputHttp(*outputHttp))
	}

	// 设置UDP转发的目标地址
	if *outputUdp != "" {
		options = append(options, mirror.WithOutputUdp(*outputUdp))
	}

	if *debug == true {
		log.SetLevel(log.DebugLevel)
		log.SetFormatter(&log.TextFormatter{
			DisableColors: false,
			FullTimestamp: true,
		})
	} else {
		log.SetLevel(log.InfoLevel)
		log.SetFormatter(&log.TextFormatter{
			DisableColors: false,
			FullTimestamp: true,
		})
	}

	packetmirror := mirror.New(options...)

	// 启动服务
	return packetmirror.Start()
}
