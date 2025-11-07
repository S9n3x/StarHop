package general

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"os"
	"runtime"
	"strings"
	"sync"
)

var (
	DeviceIDOnce   sync.Once
	cachedDeviceID string
)

// GenerateDeviceID 0x + 指纹哈希前8位
// 指纹组合：优先物理网卡 MAC + 主机名 + OS + ARCH；无物理 MAC 则忽略。
func GenerateDeviceID() string {
	DeviceIDOnce.Do(func() {
		host := getHostname()
		mac := getPreferredMAC()
		parts := []string{}
		if mac != "" {
			parts = append(parts, "mac:"+mac)
		}
		parts = append(parts, "host:"+host, "os:"+runtime.GOOS, "arch:"+runtime.GOARCH)
		fp := strings.Join(parts, "|")
		sum := sha256.Sum256([]byte(fp))
		short := hex.EncodeToString(sum[:4])
		cachedDeviceID = "0x" + short
	})
	return cachedDeviceID
}

// getHostname 获取主机名
func getHostname() string {
	h, err := os.Hostname()
	if err != nil || h == "" {
		return "unknown"
	}
	return h
}

// getPreferredMAC 获取一个合适的物理网卡 MAC（过滤回环/虚拟/未启用）
func getPreferredMAC() string {
	ifs, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifs {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		nameLower := strings.ToLower(iface.Name)
		if strings.HasPrefix(nameLower, "docker") ||
			strings.HasPrefix(nameLower, "veth") ||
			strings.HasPrefix(nameLower, "br-") ||
			strings.HasPrefix(nameLower, "lo") {
			continue
		}
		if len(iface.HardwareAddr) == 0 {
			continue
		}
		return iface.HardwareAddr.String()
	}
	return ""
}
