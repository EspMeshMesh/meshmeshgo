package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"leguru.net/m/v2/logger"
)

type Config struct {
	WantHelp           bool
	ConfigFile         string
	SerialPortName     string `json:"SerialPortName"`
	SerialPortBaudRate int    `json:"SerialPortBaudRate"`
	SerialIsEsp8266    bool   `json:"SerialIsEsp8266"`
	SerialShouldRetry  bool		`json:"SerialShouldRetry"`
	VerboseLevel       int    `json:"VerboseLevel"`
	TargetNode         int    `json:"TargetNode"`
	DebugNodeAddr      string `json:"DebugNodeAddr"`
	RestBindAddress    string `json:"RestBindAddress"`
	RpcBindAddress     string `json:"RpcBindAddress"`
	BindAddress        string `json:"BindAddress"`
	BindPort           int    `json:"BindPort"`
	BasePortOffset     int    `json:"BasePortOffset"`
	SizeOfPortsPool    int    `json:"SizeOfPortsPool"`
	EnableZeroconf     bool   `json:"EnableZeroconf"`
}

func NewConfig() (*Config, error) {
	var err error
	config := Config{
		WantHelp:           true,
		VerboseLevel:       0,
		RestBindAddress:    ":4040",
		BindAddress:        "dynamic",
		BindPort:           6053,
		BasePortOffset:     20000,
		SizeOfPortsPool:    10000,
		ConfigFile:         "meshmeshgo.json",
		SerialPortName:     "/dev/ttyUSB0",
		SerialPortBaudRate: 460800,
		SerialShouldRetry:  true,
		EnableZeroconf:     false,
	}

	app := &cli.App{
		Name:  "meshmeshgo",
		Usage: "meshmesh hub written in go!",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "port",
				Value:       config.SerialPortName,
				Usage:       "Serial port name",
				Aliases:     []string{"p"},
				Destination: &config.SerialPortName,
			},
			&cli.IntFlag{
				Name:        "baud",
				Value:       config.SerialPortBaudRate,
				Aliases:     []string{"b"},
				Destination: &config.SerialPortBaudRate,
			},
			&cli.BoolFlag{
				Name:        "esp8266",
				Value:       false,
				Usage:       "Set if the coordinator is an esp8266",
				Destination: &config.SerialIsEsp8266,
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Count:   &config.VerboseLevel,
			},
			&cli.IntFlag{
				Name:        "target",
				Aliases:     []string{"t"},
				Destination: &config.TargetNode,
				Base:        16,
			},
			&cli.StringFlag{
				Name:        "node_to_debug",
				Aliases:     []string{"dbg"},
				Usage:       "Debug a single node connection",
				Destination: &config.DebugNodeAddr,
			},
			&cli.StringFlag{
				Name:        "dashboard",
				Value:       config.RestBindAddress,
				Aliases:     []string{"d"},
				Usage:       "Bind address for the dashboard server",
				Destination: &config.RestBindAddress,
			},
			&cli.StringFlag{
				Name:        "rpc_bind_address",
				Value:       config.RpcBindAddress,
				Usage:       "Bind address for the rpc server",
				Destination: &config.RpcBindAddress,
			},
			&cli.StringFlag{
				Name:        "bind_address",
				Value:       config.BindAddress,
				Usage:       "Bind address for the esphome servers. Use 'dynamic' to auto-assign a port based on the node id",
				Destination: &config.BindAddress,
			},
			&cli.IntFlag{
				Name:        "bind_port",
				Value:       config.BindPort,
				Usage:       "Bind port for the esphome servers. Use 0 to auto-assign a port based on the bind address",
				Destination: &config.BindPort,
			},
			&cli.IntFlag{
				Name:        "base_port_offset",
				Value:       config.BasePortOffset,
				Usage:       "Base port offset for the esphome servers",
				Destination: &config.BasePortOffset,
			},
			&cli.IntFlag{
				Name:        "size_of_ports_pool",
				Value:       config.SizeOfPortsPool,
				Usage:       "Size of ports pool for the server",
				Destination: &config.SizeOfPortsPool,
			},
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Value:       config.ConfigFile,
				Destination: &config.ConfigFile,
			},
		},
		Action: func(cCtx *cli.Context) error {
			config.WantHelp = false
			return nil
		},
	}

	if err = app.Run(os.Args); err != nil {
		logger.Log().Fatal(err)
	}

	if _, err = os.Stat(config.ConfigFile); err == nil {
		data, err := os.ReadFile(config.ConfigFile)
		fmt.Println("Data: " + string(data))
		if err == nil {
			json.Unmarshal(data, &config)
		}
	}
	if err != nil {
		res2B, _ := json.MarshalIndent(&config, "", "  ")
		fmt.Println(string(res2B))
		logger.Log().Fatal(err)
	}

	return &config, err
}
