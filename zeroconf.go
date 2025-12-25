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

	srv.Text = map[string]string{
		"friendly_name":   name,
		"mac":             "FE7F00" + fmt.Sprintf("%06X", nodeid&0xFFFFFF),
		"board":           "esp32dev",
		"project_name":    name,
		"project_version": "1.0.1",
		"network":         "meshmesh",
		"version":         firmware,
	}

	h, err := z.rp.Add(srv, true)
	if err != nil {
		return err
	}

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

func (z *ZeroconfResponder) networkChangedCallback(network *graph.Network) {
	nodes := network.Nodes()
	for nodes.Next() {
		node := nodes.Node().(graph.NodeDevice)
		wantService := node.Device().InUse() && !network.IsLocalDevice(node) && !node.Device().DeepSleep()
		hasService := z.services[node.DeviceTagOrFormattedId()] != nil

		if wantService && !hasService {
			port := utils.ComputeNodePort(node.ID(), 6053, 20000, 10000)
			z.addService(node.DeviceTagOrFormattedId(), int32(node.ID()), port, node.Device().Firmware())
			logger.WithFields(logger.Fields{"node": node.DeviceTagOrFormattedId(), "port": port}).Info("ZeroconfResponder.Adding Zeroconf service")
		} else if !wantService && hasService {
			logger.WithFields(logger.Fields{"node": node.DeviceTagOrFormattedId()}).Info("ZeroconfResponder.networkChangedCallback: Removing Zeroconf service")
			z.removeService(node.DeviceTagOrFormattedId())
		}
	}
}

func (z *ZeroconfResponder) Start(network *graph.Network) error {
	err := z.setupZeroconf()
	if err != nil {
		return err
	}

	network.AddNetworkChangedCallback(z.networkChangedCallback)
	z.networkChangedCallback(network)

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
