package browser

import (
	"fmt"
	"net"
)

func FindAvailablePort(startPort, endPort int) (int, error) {
	for port := startPort; port <= endPort; port++ {
		addr := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			listener.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("端口 %d-%d 范围内无可用端口", startPort, endPort)
}
