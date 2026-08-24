package server

import (
	"fmt"
	"net"
	"strings"
)

func IsLoopbackAddress(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return strings.HasPrefix(addr, "127.")
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
func ValidatePort(addr string) error {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("地址必须包含端口")
	}
	if port == "" {
		return fmt.Errorf("端口不能为空")
	}
	return nil
}
func StartupMessage(addr string) string {
	if IsLoopbackAddress(addr) {
		return fmt.Sprintf("工作台已准备，在 http://%s/workbench 访问", addr)
	}
	return fmt.Sprintf("工作台监听 %s", addr)
}
func ConfigSummary(c Config) map[string]string {
	return map[string]string{"addr": c.Addr, "dataDir": c.DataDir, "mode": map[bool]string{true: "selfcheck", false: "server"}[c.SelfCheck]}
}
