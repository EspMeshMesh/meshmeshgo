package main

import (
	"fmt"
	"net"

	"github.com/hashicorp/mdns"
	"github.com/miekg/dns"
	gra "leguru.net/m/v2/graph"
	"leguru.net/m/v2/logger"
	"leguru.net/m/v2/utils"
)

type MdnsServiceConfig struct {
	zeroconfEnabled bool
	dynmicAddress   bool
	apiPort         int
	apiBasePort     int
	apiPortsSpan    int
}

type MultiMdnsService struct {
	services []*mdns.MDNSService
}

var mdnsServers []*mdns.Server = nil
var mdnsConfig MdnsServiceConfig = MdnsServiceConfig{
	zeroconfEnabled: false,
	dynmicAddress:   true,
	apiPort:         6053,
	apiBasePort:     20000,
	apiPortsSpan:    10000,
}

func setMdnsConfig(config MdnsServiceConfig) {
	mdnsConfig = config
}

func (m *MultiMdnsService) AddService(service *mdns.MDNSService) {
	if service == nil {
		panic("provided service is nil")
	}
	m.services = append(m.services, service)
}

func (m *MultiMdnsService) Records(q dns.Question) []dns.RR {
	records := make([]dns.RR, 0)
	for _, service := range m.services {
		records = append(records, service.Records(q)...)
	}
	return records
}

func NewMultiMdnsService() *MultiMdnsService {
	return &MultiMdnsService{services: make([]*mdns.MDNSService, 0)}
}

func setupMdns() {
	if !mdnsConfig.zeroconfEnabled {
		return
	}

	if mdnsServers != nil {
		for _, mdnsServer := range mdnsServers {
			mdnsServer.Shutdown()
		}
		mdnsServers = nil
	}

	mdnsServers = make([]*mdns.Server, 0)

	network := gra.GetMainNetwork()
	nodes := network.Nodes()
	//multiService := NewMultiMdnsService()
	for nodes.Next() {
		node := nodes.Node().(gra.NodeDevice)
		if node.Device().InUse() && !network.IsLocalDevice(node) && node.Device().Tag() != "" {
			var ips []net.IP = nil

			if mdnsConfig.dynmicAddress {
				ips = append(ips, utils.ToIPv4(node.ID()))
			}

			var port = utils.ComputeNodePort(node.ID(), mdnsConfig.apiPort, mdnsConfig.apiBasePort, mdnsConfig.apiPortsSpan)
			var firmware = node.Device().Firmware()
			if firmware == "" {
				firmware = "unknown"
			}

			service, err := mdns.NewMDNSService(
				node.DeviceTagOrFormattedId(),
				"_esphomelib._tcp",
				"local.",
				utils.ToFQDN(node.DeviceTagOrFormattedId(), "meshmesh"),
				port,
				ips,
				[]string{
					"friendly_name=" + node.Device().Tag(),
					"mac=FE7F00" + fmt.Sprintf("%06X", node.ID()),
					"board=esp32dev",
					"project_name=" + node.Device().Tag(),
					"project_version=1.0.1",
					"network=meshmesh",
					"version=" + firmware,
				},
			)
			if err != nil {
				panic(err)
			}
			//multiService.AddService(service)
			logger.WithFields(logger.Fields{"node": node.Device().Tag(), "port": port, "dynamic": mdnsConfig.dynmicAddress}).Info("Adding mDNS service")
			mdnsServer, err := mdns.NewServer(&mdns.Config{Zone: service})
			if err != nil {
				panic(err)
			}
			mdnsServers = append(mdnsServers, mdnsServer)
		}
	}
}
