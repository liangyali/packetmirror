package cmd

import (
	"flag"

	"github.com/liangyali/packetmirror/internal/mirror"
)

func Run() error {

	device := flag.String("device", "", "网卡名称")
	inputFilter := flag.String("input-filter", "", "流量过滤条件")
	outputHttp := flag.String("output-http", "http://127.0.0.1:80", "http转发目标地址.")
	outputUdp := flag.String("output-udp", "127.0.0.1:80", "udp转发目标地址.")

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

	packetmirror := mirror.New(options...)

	// 启动服务
	return packetmirror.Start()
}
