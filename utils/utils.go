package utils

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func FmtNodeId(nodeid int64) string {
	return fmt.Sprintf("0x%06X", nodeid)
}

func FmtNodeIdHass(nodeid int64) string {
	return fmt.Sprintf("127.%d.%d.%d", (nodeid>>16)&0xFF, (nodeid>>8)&0xFF, nodeid&0xFF)
}

func ForceDebug(force bool, data interface{}) {
	/*var level logrus.Level = logrus.DebugLevel
	if force {
		level = logrus.InfoLevel
	}
	logrus.(level, data)*/
}

func ForceDebugEntry(entry *logrus.Entry, force bool, data interface{}) {
	var level logrus.Level = logrus.DebugLevel
	if force {
		level = logrus.InfoLevel
	}
	entry.Log(level, data)
}
