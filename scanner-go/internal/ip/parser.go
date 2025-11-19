package ip

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

// ParseAny — принимает CIDR, диапазон "A-B" или одиночный IP.
func ParseAny(input string) ([]string, error) {
	// CIDR
	if strings.Contains(input, "/") {
		return fromCIDR(input)
	}

	// Диапазон IP
	if strings.Contains(input, "-") {
		return fromRange(input)
	}

	// Одиночный IP
	if net.ParseIP(input) != nil {
		return []string{input}, nil
	}

	return nil, errors.New("Неверный формат Ip")
}

// CIDR
func fromCIDR(cidr string) ([]string, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ipnet.IP.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	if len(ips) > 2 {
		ips = ips[1 : len(ips)-1] // убираем network/broadcast
	}

	return ips, nil
}

// диапозон
func fromRange(r string) ([]string, error) {
	parts := strings.Split(r, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Неверный диапозон: %s", r)
	}

	start := net.ParseIP(parts[0])
	end := net.ParseIP(parts[1])
	if start == nil || end == nil {
		return nil, errors.New("Неверный IP в диапозоне")
	}

	var ips []string
	for ip := start; !ipEqual(ip, end); inc(ip) {
		ips = append(ips, ip.String())
	}
	ips = append(ips, end.String())

	return ips, nil
}

// Помошники
func ipEqual(a, b net.IP) bool {
	return a.String() == b.String()
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] != 0 {
			break
		}
	}
}
