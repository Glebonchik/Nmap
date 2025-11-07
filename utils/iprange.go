package util

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

// IPsFromCIDR возвращает список IP-адресов для заданного CIDR.
func IPsFromCIDR(cidr string) ([]string, error) {
ip, ipnet, err := net.ParseCIDR(cidr)
if err != nil {
return nil, err
}


var ips []string
for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
ips = append(ips, ip.String())
}


// Удалим сеть и широковещательный адрес при наличии
if len(ips) > 2 {
ips = ips[1 : len(ips)-1]
}
return ips, nil
}


// IPsFromRange парсит строку вида "start-end" и возвращает все IP в диапазоне
func IPsFromRange(r string) ([]string, error) {
parts := strings.Split(r, "-")
if len(parts) != 2 {
return nil, errors.New("range must be start-end")
}
start := net.ParseIP(strings.TrimSpace(parts[0]))
end := net.ParseIP(strings.TrimSpace(parts[1]))
if start == nil || end == nil {
return nil, fmt.Errorf("invalid ip address")
}


if !sameV4(start, end) {
return nil, fmt.Errorf("only IPv4 ranges supported in this util")
}


// инкрементируем
var ips []string
for ip := start.To4(); ; inc(ip) {
ips = append(ips, ip.String())
if ip.Equal(end) {
break
}
}
return ips, nil
}


func sameV4(a, b net.IP) bool {
if a.To4() == nil || b.To4() == nil {
return false
}
return true
}


// inc увеличивает IP (in-place)
func inc(ip net.IP) {
for j := len(ip) - 1; j >= 0; j-- {
ip[j]++
if ip[j] > 0 {
break
}
}
}