package utils

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
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

func ToIPv4(ip int64) net.IP {
	return net.IPv4(127, byte((ip>>16)&0xFF), byte((ip>>8)&0xFF), byte(ip&0xFF))
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

func TruncateZeros(s []byte) string {
	pos := bytes.IndexByte(s, 0)
	if pos == -1 {
		return string(s)
	}
	return string(s[:bytes.IndexByte(s, 0)])
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

func ComputeNodePort(nodeid int64, port int, base int, span int) int {
	if port > 0 {
		return port
	}
	return HashString(FmtNodeId(nodeid), span) + base
}

func ToFQDN(tag string, domain string) string {
	if tag == "" {
		tag = "unknow"
	}
	str := strings.ToLower(tag)
	str = strings.Replace(str, " ", "_", -1)
	str = strings.Replace(str, ".", "_", -1)
	return str + "." + domain + "."
}

func RevisionToInteger(revision string) int {
	if strings.Contains(revision, ",") {
		return 0
	}
	parts := strings.Split(revision, ".")
	if len(parts) != 3 {
		return 0
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0
	}
	return major*1000000 + minor*1000 + patch
}
