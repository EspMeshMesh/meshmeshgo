package main

import (
	"context"
	"fmt"
	"net"

	"github.com/brutella/dnssd"
	"leguru.net/m/v2/graph"
	"leguru.net/m/v2/logger"
	"leguru.net/m/v2/utils"
)

type ZeroconfResponder struct {
	ctx      context.Context
	cancel   context.CancelFunc
	rp       dnssd.Responder
	services map[string]dnssd.ServiceHandle
}

func (z *ZeroconfResponder) setupZeroconf() error {
	var err error
	z.rp, err = dnssd.NewResponder()
	return err
}

func (z *ZeroconfResponder) addService(name string, nodeid int32, port int, firmware string) error {
	cfg := dnssd.Config{
		Name: name,
		Type: "_esphomelib._tcp",
		Port: port,
		Host: utils.ToFQDN(name, "meshmesh"),
		IPs:  []net.IP{utils.ToIPv4(int64(nodeid))},
	}

	srv, err := dnssd.NewService(cfg)
	if err != nil {
		return err
	}

	h, err := z.rp.Add(srv)
	if err != nil {
		return err
	}

	h.UpdateText(
		map[string]string{
			"friendly_name":   name,
			"mac":             "FE7F00" + fmt.Sprintf("%06X", nodeid&0xFFFFFF),
			"board":           "esp32dev",
			"project_name":    name,
			"project_version": "1.0.1",
			"network":         "meshmesh",
			"version":         firmware,
		}, z.rp,
	)

	z.services[name] = h
	return nil
}

func (z *ZeroconfResponder) removeService(name string) error {
	h, ok := z.services[name]
	if !ok {
		return fmt.Errorf("service not found")
	}
	z.rp.Remove(h)
	delete(z.services, name)
	return nil
}

func (z *ZeroconfResponder) networkChangedCallback(network *graph.Network) error {
	nodes := network.Nodes()
	for nodes.Next() {
		node := nodes.Node().(graph.NodeDevice)
		wantService := node.Device().InUse() && !network.IsLocalDevice(node)
		hasService := z.services[node.DeviceTagOrFormattedId()] != nil

		if wantService && !hasService {
			logger.WithFields(logger.Fields{"node": node.DeviceTagOrFormattedId()}).Info("Adding Zeroconf service")
			port := utils.ComputeNodePort(node.ID(), 6053, 20000, 10000)
			z.addService(node.DeviceTagOrFormattedId(), int32(node.ID()), port, node.Device().Firmware())
		} else if !wantService && hasService {
			logger.WithFields(logger.Fields{"node": node.DeviceTagOrFormattedId()}).Info("Removing Zeroconf service")
			z.removeService(node.DeviceTagOrFormattedId())
		}
	}
	return nil
}

func (z *ZeroconfResponder) Start(network *graph.Network) error {
	err := z.setupZeroconf()
	if err != nil {
		return err
	}

	nodes := network.Nodes()
	for nodes.Next() {
		node := nodes.Node().(graph.NodeDevice)
		if node.Device().InUse() && !network.IsLocalDevice(node) {
			logger.WithFields(logger.Fields{"node": node.DeviceTagOrFormattedId()}).Info("Adding Zeroconf service")
			port := utils.ComputeNodePort(node.ID(), 6053, 20000, 10000)
			z.addService(node.DeviceTagOrFormattedId(), int32(node.ID()), port, node.Device().Firmware())
		}
	}

	network.AddNetworkChangedCallback(func(network *graph.Network) {
		z.networkChangedCallback(network)
	})

	z.ctx, z.cancel = context.WithCancel(context.Background())
	go z.rp.Respond(z.ctx)
	return nil
}

func (z *ZeroconfResponder) Stop() {
	for _, service := range z.services {
		z.rp.Remove(service)
	}
	z.services = make(map[string]dnssd.ServiceHandle)
	z.cancel()
}

func NewZeroconfResponder() ZeroconfResponder {
	return ZeroconfResponder{rp: nil, services: make(map[string]dnssd.ServiceHandle)}
}
