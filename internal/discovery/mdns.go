package discovery

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/mdns"
)

const serviceType = "_machine-agent._tcp"

func Advertise(port int) (func(), error) {
	hostname, _ := os.Hostname()

	info := []string{
		fmt.Sprintf("machine-agent on %s", hostname),
	}

	service, err := mdns.NewMDNSService(hostname, serviceType, "", "", port, nil, info)
	if err != nil {
		return nil, fmt.Errorf("mdns service: %w", err)
	}

	server, err := mdns.NewServer(&mdns.Config{Zone: service})
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
	params.Entries = entriesCh

	if err := mdns.Query(params); err != nil {
		return nil, fmt.Errorf("mdns query: %w", err)
	}

	return entries, nil
}
