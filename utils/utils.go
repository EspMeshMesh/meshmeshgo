package utils

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"hash/fnv"

	"github.com/sirupsen/logrus"
)

func FmtNodeId(nodeid int64) string {
	if nodeid == 0 {
		return ""
	}

	return fmt.Sprintf("N%06X", nodeid)
}

func ParseNodeId(id any) (int64, error) {
	switch id := id.(type) {
	case string:
		if len(id) < 1 {
			return 0, errors.New("invalid id string")
		}
		id = strings.Replace(id, "N", "0x", 1)
		return strconv.ParseInt(id, 0, 32)
	default:
		return -1, errors.New("invalid id string")
	}
}

func FmtNodeIdHass(nodeid int64) string {
	return fmt.Sprintf("127.%d.%d.%d", (nodeid>>16)&0xFF, (nodeid>>8)&0xFF, nodeid&0xFF)
}

func FmtPath2Str(path []int64) string {
	var _path string
	for _, p := range path {
		if len(_path) > 0 {
			_path += " > "
		}
		_path += FmtNodeId(p)
	}
	return _path
}

func ForceDebugEntry(entry *logrus.Entry, force bool, data interface{}) {
	var level logrus.Level = logrus.DebugLevel
	if force {
		level = logrus.InfoLevel
	}
	entry.Log(level, data)
}

func EncodeToHexEllipsis(data []byte, maxlen int) string {
	str := hex.EncodeToString(data[0:min(len(data), maxlen)])
	if len(data) > maxlen {
		str += "..."
	}
	return str
}

func HashString(s string, mod int) int {
	hash := fnv.New32()
	hash.Write([]byte(s))
	hashValue := hash.Sum32()
	hashValue = hashValue % uint32(mod)
	return int(hashValue)
}

func FindFirstZeroChar(s []byte) int {
	for i, c := range s {
		if c == 0 {
			return i
		}
	}
	return len(s)
}

func TruncateZeros(s []byte) string {
	return string(s[:FindFirstZeroChar(s)])
}

func BackupFile(filename string, backupdir string) {
	if _, err := os.Stat(backupdir); err != nil {
		os.MkdirAll(backupdir, 0755)
	}
	ext := filepath.Ext(filename)
	filenamenoext := strings.TrimSuffix(filename, ext)
	backupfile := filenamenoext + "_" + time.Now().Format("20060102150405") + ext + ".bak"
	if _, err := os.Stat(filename); err == nil {
		os.Rename(filename, filepath.Join(backupdir, backupfile))
	}
}
