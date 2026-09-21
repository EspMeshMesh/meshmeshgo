package meshmesh

import (
	"bytes"
	"fmt"
	"time"

	"leguru.net/m/v2/logger"
	"leguru.net/m/v2/utils"
)

const (
	otaChunkSize = 512

	otaBaseDelayMs   = 50
	otaPerHopDelayMs = 50
)

type OtaConnection struct {
}

func (client *OtaConnection) Socket2Serial(buffer *bytes.Buffer, connectedPath *ConnPathConnection, stats *EspApiConnectionStats) {
	if buffer.Len() > 0 {
		logger.WithFields(logger.Fields{"handle": connectedPath.handle, "len": buffer.Len()}).
			Trace(fmt.Sprintf("flushBuffer: HA-->SE: %s", utils.EncodeToHexEllipsis(buffer.Bytes(), 32)))

		pathHops := connectedPath.GetPathHops()
		logger.WithFields(logger.Fields{"handle": connectedPath.handle, "len": buffer.Len(), "pathHops": pathHops}).
			Info("OtaConnection.Socket2Serial")

		chunks := (buffer.Len()-1)/otaChunkSize + 1
		for i := range chunks {
			chunk := buffer.Next(otaChunkSize)
			err := connectedPath.SendData(chunk)
			if err != nil {
				logger.Log().Error(fmt.Sprintf("Error writing on serial: %s", err.Error()))
				break
			}

			sleepTimeUs := client.calculatePacingDelay(connectedPath, len(chunk), pathHops)
			if sleepTimeUs > 0 {
				logger.WithFields(logger.Fields{
					"handle":   connectedPath.handle,
					"chunk":    i + 1,
					"total":    chunks,
					"delayMs":  sleepTimeUs / 1000,
					"pathHops": pathHops,
				}).Trace("OtaConnection.Socket2Serial: pacing delay")
				time.Sleep(time.Duration(sleepTimeUs) * time.Microsecond)
			}
		}

		stats.SentBytes(buffer.Len())
		buffer.Reset()
	}
}

func (client *OtaConnection) calculatePacingDelay(connectedPath *ConnPathConnection, chunkLen int, pathHops int) int {
	if connectedPath.serialProxy.IsEsp8266() {
		sleepTime := connectedPath.serialProxy.TxOneByteMs() * (chunkLen * 2)
		if sleepTime < 150000 {
			sleepTime = 150000
		}
		return sleepTime
	}

	if pathHops > 1 {
		delayMs := otaBaseDelayMs + (pathHops * otaPerHopDelayMs)
		return delayMs * 1000
	}

	return 0
}

func NewOtaConnection() *OtaConnection {
	return &OtaConnection{}
}
