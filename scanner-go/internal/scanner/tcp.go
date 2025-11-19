package scanner

import (
	"fmt"
	"net"
	"time"
)

type PortResult struct {
	Port int  `json:"port"`
	Open bool `json:"open"`
}

type HostScan struct {
	IP    string       `json:"ip"`
	Ports []PortResult `json:"ports"`
}

// Сканирует все порты из списка ports
func ScanPorts(ip string, ports []int, timeout time.Duration) HostScan {
	results := HostScan{
		IP:    ip,
		Ports: make([]PortResult, 0, len(ports)),
	}

	for _, port := range ports {
		open := scanSinglePort(ip, port, timeout)
		results.Ports = append(results.Ports, PortResult{
			Port: port,
			Open: open,
		})
	}

	return results
}

// Одиночная проверка порта
func scanSinglePort(ip string, port int, timeout time.Duration) bool {
	addr := net.JoinHostPort(ip, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", addr, timeout)

	if err != nil {
		return false
	}

	conn.Close()
	return true
}
