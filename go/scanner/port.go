package scanner

import (
	"fmt"
	"net"
	"time"
)

// Проверка одного TCP-порта
func CheckPort(ip string, port int, timeout int) bool {
	address := net.JoinHostPort(ip, fmt.Sprintf("%d", port))

	conn, err := net.DialTimeout("tcp", address, time.Duration(timeout)*time.Millisecond)
	if err != nil {
		return false
	}

	_ = conn.Close()
	return true
}
