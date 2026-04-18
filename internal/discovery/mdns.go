package discovery

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/hashicorp/mdns"
)

var silentLogger = log.New(io.Discard, "", 0)

const serviceType = "_machine-agent._tcp"

func Advertise(port int) (func(), error) {
	hostname, _ := os.Hostname()

	info := []string{
		fmt.Sprintf("machine-agent on %s", hostname),
	}

	instance := fmt.Sprintf("%s-%d", hostname, port)
	service, err := mdns.NewMDNSService(instance, serviceType, "", "", port, nil, info)
	if err != nil {
		return nil, fmt.Errorf("mdns service: %w", err)
	}

	server, err := mdns.NewServer(&mdns.Config{Zone: service, Logger: silentLogger})
	if err != nil {
		return nil, fmt.Errorf("mdns server: %w", err)
	}

	log.Printf("mDNS: advertising %s on %s:%d", serviceType, hostname, port)

	return func() {
		server.Shutdown()
	}, nil
}

func Discover() ([]*mdns.ServiceEntry, error) {
	var entries []*mdns.ServiceEntry

	entriesCh := make(chan *mdns.ServiceEntry, 16)

	go func() {
		for entry := range entriesCh {
			entries = append(entries, entry)
		}
	}()

	params := mdns.DefaultParams(serviceType)
	params.DisableIPv6 = true
	params.Timeout = 3 * time.Second
	params.Entries = entriesCh
	params.Logger = silentLogger

	if err := mdns.Query(params); err != nil {
		return nil, fmt.Errorf("mdns query: %w", err)
	}

	return entries, nil
}
