package processor

import (
	"fmt"
	"sync"

	"github.com/google/gopacket"
	"github.com/liangyali/packetmirror/settings"
)

var (
	rwLock     sync.RWMutex
	processors = make(map[string]PacketProcessor)
)

// PacketProcessor 解析协议
type PacketProcessor interface {

	// Process 处理网路协议包
	Process(packet gopacket.Packet, settings settings.Settings)
}

// Register makes a processor available by the provided name.
func Register(name string, processor PacketProcessor) {
	rwLock.Lock()
	defer rwLock.Unlock()

	if processor == nil {
		panic("processor: Register driver is nil")
	}
	if _, dup := processors[name]; dup {
		panic("processor: Register called twice for processor " + name)
	}

	processors[name] = processor
}

func Process(packet gopacket.Packet, settings settings.Settings) {

	applicationLayer := packet.ApplicationLayer()
	if applicationLayer != nil {
		processor, ok := processors["http"]
		if ok {
			processor.Process(packet, settings)
		} else {
			fmt.Println("没有找到合适httpforward处理器")
		}
	}
}
