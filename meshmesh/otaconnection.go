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

	// Multi-hop mesh pacing delays for OTA bulk transfer.
	// These delays STACK with ESP8266 serial pacing when both apply.
	// Tuned for reliable delivery: each hop adds radio TX time, MAC retries,
	// and relay processing. Conservative values prevent buffer overflow in
	// intermediate mesh nodes.
	otaMultiHopBaseDelayMs = 200 // Base delay for any multi-hop path (mesh overhead)
	otaPerHopDelayMs       = 200 // Additional delay per hop (radio relay time)

	// Minimum ESP8266 serial pacing delay in microseconds
	otaEsp8266MinDelayUs = 150000
)

type OtaConnection struct {
}

func (client *OtaConnection) Socket2Serial(buffer *bytes.Buffer, connectedPath *ConnPathConnection, stats *EspApiConnectionStats) {
	if buffer.Len() > 0 {
		logger.WithFields(logger.Fields{"handle": connectedPath.handle, "len": buffer.Len()}).
			Trace(fmt.Sprintf("flushBuffer: HA-->SE: %s", utils.EncodeToHexEllipsis(buffer.Bytes(), 32)))

		pathHops := connectedPath.GetPathHops()
		isEsp8266 := connectedPath.serialProxy.IsEsp8266()

		logger.WithFields(logger.Fields{
			"handle":    connectedPath.handle,
			"len":       buffer.Len(),
			"pathHops":  pathHops,
			"isEsp8266": isEsp8266,
		}).Info("OtaConnection.Socket2Serial")

		chunks := (buffer.Len()-1)/otaChunkSize + 1
		for i := range chunks {
			chunk := buffer.Next(otaChunkSize)
			err := connectedPath.SendData(chunk)
			if err != nil {
				logger.Log().Error(fmt.Sprintf("Error writing on serial: %s", err.Error()))
				break
			}

			esp8266DelayUs, hopDelayUs, totalDelayUs := client.calculatePacingDelay(connectedPath, len(chunk), pathHops)
			if totalDelayUs > 0 {
				logger.WithFields(logger.Fields{
					"handle":       connectedPath.handle,
					"chunk":        i + 1,
					"total":        chunks,
					"esp8266DelMs": esp8266DelayUs / 1000,
					"hopDelMs":     hopDelayUs / 1000,
					"totalDelMs":   totalDelayUs / 1000,
					"pathHops":     pathHops,
				}).Trace("OtaConnection.Socket2Serial: pacing delay breakdown")
				time.Sleep(time.Duration(totalDelayUs) * time.Microsecond)
			}
		}

		stats.SentBytes(buffer.Len())
		buffer.Reset()
	}
}

// calculatePacingDelay computes the total pacing delay for an OTA chunk.
// Returns (esp8266DelayUs, hopDelayUs, totalDelayUs).
// ESP8266 serial pacing and multi-hop mesh pacing are STACKED (summed),
// not mutually exclusive. This ensures multi-hop paths get adequate delay
// even when the coordinator uses ESP8266 serial timing.
func (client *OtaConnection) calculatePacingDelay(connectedPath *ConnPathConnection, chunkLen int, pathHops int) (int, int, int) {
	var esp8266DelayUs int
	var hopDelayUs int

	// ESP8266 serial pacing: based on TX byte timing with minimum floor
	if connectedPath.serialProxy.IsEsp8266() {
		esp8266DelayUs = connectedPath.serialProxy.TxOneByteMs() * (chunkLen * 2)
		if esp8266DelayUs < otaEsp8266MinDelayUs {
			esp8266DelayUs = otaEsp8266MinDelayUs
		}
	}

	// Multi-hop mesh pacing: scaled by number of hops
	// Applied for any path with more than 1 hop (i.e., not direct neighbor)
	if pathHops > 1 {
		hopDelayMs := otaMultiHopBaseDelayMs + (pathHops * otaPerHopDelayMs)
		hopDelayUs = hopDelayMs * 1000
	}

	// Stack both delays - they address different bottlenecks:
	// - ESP8266 delay: serial TX buffer drain time
	// - Hop delay: mesh network relay and radio transmission time
	totalDelayUs := esp8266DelayUs + hopDelayUs

	return esp8266DelayUs, hopDelayUs, totalDelayUs
}

func NewOtaConnection() *OtaConnection {
	return &OtaConnection{}
}
