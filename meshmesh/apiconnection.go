package meshmesh

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"leguru.net/m/v2/logger"
)

type ApiConnection struct {
}

func (client *ApiConnection) Socket2Serial(buffer *bytes.Buffer, connectedPath *ConnPathConnection, stats *EspApiConnectionStats) {
	logger.WithField("handle", connectedPath.handle).Trace(fmt.Sprintf("HA-->SE: %s", hex.EncodeToString(buffer.Bytes())))
	err := connectedPath.SendData(buffer.Bytes())
	stats.SentBytes(buffer.Len())
	if err != nil {
		logger.Log().Error(fmt.Sprintf("Error writng on socket: %s", err.Error()))
	}

	buffer.Reset()
}

func NewApiConnection() *ApiConnection {
	return &ApiConnection{}
}
