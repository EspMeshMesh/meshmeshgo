package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"leguru.net/m/v2/config"
	gra "leguru.net/m/v2/graph"
	"leguru.net/m/v2/logger"
	"leguru.net/m/v2/meshmesh"
	"leguru.net/m/v2/rest"
	"leguru.net/m/v2/rpc"
	"leguru.net/m/v2/utils"
)

const (
	programName        = "meshmeshgo"
	programDescription = "hub server for meshmesh network"
	graphFilename      = "meshmesh.graphml"
)

var (
	vcsHash  string
	vcsTime  time.Time
	vcsDirty bool
)

var quitProgram bool = false
var debugNodeId gra.NodeDevice

func waitForTermination() {
	terminationRequested := make(chan os.Signal, 1)
	signal.Notify(terminationRequested, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-terminationRequested
	logger.Info("Program termination requested")
	quitProgram = true
}

func getBuildInfo() {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		logger.Fatal("Failed to read build info")
	}

	for _, kv := range bi.Settings {
		switch kv.Key {
		case "vcs.revision":
			vcsHash = kv.Value
		case "vcs.time":
			vcsTime, _ = time.Parse(time.RFC3339, kv.Value)
		case "vcs.modified":
			vcsDirty = kv.Value == "true"
		}
	}
}

func initConfig() *config.Config {
	c, err := config.NewConfig()
	if err != nil {
		logger.Fatal("Invalid config options: ", err)
	}

	if c.WantHelp {
		os.Exit(0)
	}

	switch c.VerboseLevel {
	case 3:
		logger.SetLevel(logrus.TraceLevel)
		fmt.Println("Setting loglevel for Tracing")
	case 2:
		logger.SetLevel(logrus.DebugLevel)
		fmt.Println("Setting loglevel for Debuging")
	case 1:
		logger.SetLevel(logrus.InfoLevel)
		fmt.Println("Setting loglevel for Info")
	case 0:
		logger.SetLevel(logrus.WarnLevel)
		fmt.Println("Setting loglevel for Errors and Warning Only.")
	default:
		fmt.Printf("Setting loglevel: Unknown (%d)\n", c.VerboseLevel)
	}

	return c
}

func networkChangedCallback() {
	gra.GetMainNetwork().SaveToFile(graphFilename)
}

func initNetwork(localNodeId int64) *gra.Network {
	var network *gra.Network
	if _, err := os.Stat(graphFilename); err == nil {
		network, err = gra.NewNeworkFromFile(graphFilename, localNodeId)
		if err != nil {
			logger.Log().Fatal("Graph read error: ", err)
		}
	} else {
		network = gra.NewNetwork(localNodeId)
		network.SaveToFile(graphFilename)
	}
	return network
}

/* Initialize debug node TODO not implemented yet */
func initDebugNode(config *config.Config) {
	if len(config.DebugNodeAddr) > 0 {
		_debugNodeId, err := gra.ParseDeviceId(config.DebugNodeAddr)
		if err != nil {
			logger.WithField("err", err).Fatal("Invalid debug node id")
			return
		}
		debugNodeId, err = gra.GetMainNetwork().GetNodeDevice(_debugNodeId)
		if err != nil {
			logger.WithField("id", utils.FmtNodeId(_debugNodeId)).Fatal("Debug node not found in graph")
			return
		}
		logger.WithFields(logger.Fields{"id": debugNodeId}).Info("Enabling debug of node")
	}
}

func handleDiscAssociateReply(v *meshmesh.DiscAssociateApiReply, serialPort *meshmesh.SerialConnection) {
	network := gra.GetMainNetwork()
	logger.WithFields(logger.Fields{"server": utils.FmtNodeId(int64(v.Server)), "source": utils.FmtNodeId(int64(v.Source))}).Debug("DiscAssociateReply received")
	source, err := network.GetNodeDevice(int64(v.Source))
	if err != nil {
		source = gra.NewNodeDevice(int64(v.Source), true, "")
		network.AddNode(source)
	}
	for i := range 3 {
		if v.NodeId[i] > 0 {
			node, err := network.GetNodeDevice(int64(v.NodeId[i]))
			if err != nil {
				network.ChangeEdgeWeight(node.ID(), source.ID(), meshmesh.Rssi2weight(v.Rssi[i]), meshmesh.Rssi2weight(v.Rssi[i]))
				logger.WithFields(logger.Fields{"id": utils.FmtNodeId(int64(v.NodeId[i])), "rssi": v.Rssi[i]}).Debug("DiscAssociateReply received")
			}
		}
	}
	network.SaveToFile(graphFilename)
	serialPort.SendReceiveApiProt(meshmesh.NodeIdApiRequest{}, meshmesh.UnicastProtocol, meshmesh.MeshNodeId(source.ID()), nil)
	// ***** TODO: Update network graph with new node
}

// @title           Meshmesh API
// @version         1.0.0
// @description     Meshmesh API documents https://github.com/EspMeshMesh/meshmeshgo
// @termsOfService  http://swagger.io/terms/

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @contact.name    Meshmesh
// @contact.url     https://github.com/EspMeshMesh

// @BasePath        /api/v1
// @schemes http

func main() {
	getBuildInfo()
	go waitForTermination()

	fmt.Printf("Starting program: %s\n", programName)
	fmt.Println(programDescription)
	logger.WithFields(logger.Fields{"vcsHash": vcsHash,
		"vcsTime":  vcsTime,
		"vcsDirty": vcsDirty}).Info("Startup information")

	config := initConfig()

	logger.WithFields(logger.Fields{"portName": config.SerialPortName, "baudRate": config.SerialPortBaudRate}).Debug("Opening serial port")
	// First init serial connection with coordinator

	serialPort, err := meshmesh.NewSerial(config.SerialPortName, config.SerialPortBaudRate, config.SerialIsEsp8266, false)
	if err != nil {
		logger.Log().Fatal("Serial port error: ", err)
	}
	// Init network graph
	gra.SetMainNetwork(initNetwork(int64(serialPort.LocalNode)))
	gra.AddMainNetworkChangedCallback(networkChangedCallback)
	// Init node for spcific debug
	initDebugNode(config)
	gra.PrintTable(gra.GetMainNetwork())
	// Handle DiscAssociateReply received from other nodes
	serialPort.DiscAssociateFn = handleDiscAssociateReply
	// Initialize Esphome to HomeAssistant Server
	esphomeapi := meshmesh.NewMultiServerApi(serialPort, meshmesh.ServerApiConfig{
		BindAddress:     config.BindAddress,
		BindPort:        config.BindPort,
		BasePortOffset:  config.BasePortOffset,
		SizeOfPortsPool: config.SizeOfPortsPool,
	})
	// Start RPC Server
	rpcServer := rpc.NewRpcServer(config.RpcBindAddress)
	rpcServer.Start(fmt.Sprintf("%s - %s", programName, programDescription), fmt.Sprintf("%s - %s", vcsHash, vcsTime.Format(time.RFC3339)), serialPort)
	defer rpcServer.Stop()
	// Start rest server
	restHandler := rest.NewHandler(serialPort, esphomeapi)
	rest.StartRestServer(rest.NewRouter(restHandler), config.RestBindAddress)

	var lastStatsTime time.Time
	for {
		time.Sleep(1 * time.Second)
		if quitProgram {
			break
		}
		if !serialPort.IsConnected() {
			break
		}
		if time.Since(lastStatsTime) > 1*time.Minute {
			lastStatsTime = time.Now()
			//if (len(as.Connections)> 0 ) {  //
			esphomeapi.PrintStats()
			//}
		}
	}
}
