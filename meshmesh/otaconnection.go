package meshmesh

import (
	"bytes"
	"fmt"
	"time"

	"leguru.net/m/v2/logger"
	"leguru.net/m/v2/utils"
)

type OtaConnection struct {
}

func (client *OtaConnection) Socket2Serial(buffer *bytes.Buffer, connectedPath *ConnPathConnection, stats *EspApiConnectionStats) {
	const chunkSize = 512

	if buffer.Len() > 0 {
		logger.WithFields(logger.Fields{"handle": connectedPath.handle, "len": buffer.Len()}).
			Trace(fmt.Sprintf("flushBuffer: HA-->SE: %s", utils.EncodeToHexEllipsis(buffer.Bytes(), 32)))

		chunks := (buffer.Len()-1)/chunkSize + 1
		for range chunks {
			chunk := buffer.Next(chunkSize)
			err := connectedPath.SendData(chunk)
			if err != nil {
				logger.Log().Error(fmt.Sprintf("Error writing on serial: %s", err.Error()))
				break
			}
			if connectedPath.serialProxy.IsEsp8266() {
				sleepTime := connectedPath.serialProxy.TxOneByteMs() * (len(chunk) * 2)
				if sleepTime < 150000 {
					sleepTime = 150000
				}
				time.Sleep(time.Duration(sleepTime) * time.Microsecond)
			}
		}

		stats.SentBytes(buffer.Len())
		buffer.Reset()
	}
}

func NewOtaConnection() *OtaConnection {
	return &OtaConnection{}
}
