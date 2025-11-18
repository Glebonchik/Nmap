package scanner

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"
)

// ----------------------------
// Структуры результата
// ----------------------------

type PortInfo struct {
	Port       int    `json:"port"`
	Open       bool   `json:"open"`
	Banner     string `json:"banner,omitempty"`
	TLSVersion string `json:"tls_version,omitempty"`
	TLSCipher  string `json:"tls_cipher,omitempty"`
}

type HostScan struct {
	IP    string     `json:"ip"`
	Ports []PortInfo `json:"ports"`
}

// ----------------------------
// Работа с IP
// ----------------------------

// Получить список IP из диапазона 192.168.1.1-192.168.1.50
func IPsFromRange(r string) ([]string, error) {
	parts := strings.Split(r, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("incorrect range format")
	}

	start := net.ParseIP(parts[0]).To4()
	end := net.ParseIP(parts[1]).To4()

	if start == nil || end == nil {
		return nil, fmt.Errorf("invalid IPv4 in range")
	}

	var ips []string
	for ip := start; !ipEqual(ip, end); ip = nextIP(ip) {
		ips = append(ips, ip.String())
	}
	ips = append(ips, end.String())
	return ips, nil
}

func ipEqual(a, b net.IP) bool {
	return a.String() == b.String()
}

func nextIP(ip net.IP) net.IP {
	next := make(net.IP, len(ip))
	copy(next, ip)
	for i := len(next) - 1; i >= 0; i-- {
		next[i]++
		if next[i] != 0 {
			break
		}
	}
	return next
}

// Получить список IP из CIDR — например 192.168.1.0/24
func IPsFromCIDR(cidr string) ([]string, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ipnet.IP.Mask(ipnet.Mask); ipnet.Contains(ip); ip = nextIP(ip) {
		ips = append(ips, ip.String())
	}
	return ips, nil
}

// ----------------------------
// Баннеры
// ----------------------------

// TCP баннер (например: HTTP server header)
func readTCPBanner(conn net.Conn) string {
	buf := make([]byte, 128)
	conn.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	n, _ := conn.Read(buf)
	return strings.TrimSpace(string(buf[:n]))
}

// TLS баннер (сертификат + шифр)
func readTLSBanner(ip string, port int, timeout int) (string, string) {
	addr := fmt.Sprintf("%s:%d", ip, port)

	conf := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         ip,
	}

	dialer := &net.Dialer{
		Timeout: time.Duration(timeout) * time.Millisecond,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", addr, conf)
	if err != nil {
		return "", ""
	}
	defer conn.Close()

	state := conn.ConnectionState()
	return tlsVersionString(state.Version), tls.CipherSuiteName(state.CipherSuite)
}

func tlsVersionString(v uint16) string {
	switch v {
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("TLS 0x%x", v)
	}
}

// ----------------------------
// Основной сканер
// ----------------------------

func Scan(ips []string, portList []int, timeout int) []HostScan {
	var result []HostScan

	for _, ip := range ips {
		host := HostScan{IP: ip}

		for _, port := range portList {
			info := PortInfo{
				Port: port,
				Open: false,
			}

			// Проверка порта
			address := fmt.Sprintf("%s:%d", ip, port)
			conn, err := net.DialTimeout("tcp", address, time.Duration(timeout)*time.Millisecond)
			if err != nil {
				host.Ports = append(host.Ports, info)
				continue
			}

			// Порт открыт
			info.Open = true

			// Получить TCP баннер
			banner := readTCPBanner(conn)
			info.Banner = banner

			conn.Close()

			// Попробовать TLS handshake
			tlsVer, tlsCipher := readTLSBanner(ip, port, timeout)
			if tlsVer != "" {
				info.TLSVersion = tlsVer
				info.TLSCipher = tlsCipher
			}

			host.Ports = append(host.Ports, info)
		}

		result = append(result, host)
	}

	return result
}
