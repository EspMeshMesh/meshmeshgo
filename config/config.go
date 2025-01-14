package config

import (
	"os"

	"gopkg.in/ini.v1"

	"github.com/urfave/cli/v2"
	l "leguru.net/m/v2/logger"
)

type Config struct {
	WantHelp           bool
	SerialPortName     string
	SerialPortBaudRate int
	VerboseLevel       int
	TargetNode         int
	DebugNodeAddr      string
}

var iniConfig *ini.File

func InitINIConfig() {
	var err error
	iniConfig, err = ini.Load("meshmeshgo.ini")
	if err != nil {
		iniConfig = ini.Empty()
	}
}

func GetINIValue(section string, key string) string {
	return iniConfig.Section(section).Key(key).String()
}

func SetINIValue(section string, key string, value string) {
	iniConfig.Section(section).Key(key).SetValue(value)
	iniConfig.SaveTo("meshmeshgo.ini")
}

func NewConfig() (*Config, error) {
	var err error
	config := Config{WantHelp: true, VerboseLevel: 0}

	app := &cli.App{
		Name:  "meshmeshgo",
		Usage: "meshmesh hub written in go!",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "port",
				Value:       "/dev/ttyUSB0",
				Usage:       "Serial port name",
				Destination: &config.SerialPortName,
			},
			&cli.IntFlag{
				Name:        "baud",
				Value:       460800,
				Destination: &config.SerialPortBaudRate,
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
		},
		Action: func(cCtx *cli.Context) error {
			config.WantHelp = false
			return nil
		},
	}

	if err = app.Run(os.Args); err != nil {
		l.Log().Fatal(err)
	}

	return &config, err
}
