package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func parseIPs(input string) []string {
	if strings.Contains(input, "/") {
		ips := []string{}
		ip, ipnet, _ := net.ParseCIDR(input)
		for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incIP(ip) {
			ips = append(ips, ip.String())
		}
		return ips
	}

	if strings.Contains(input, "-") {
		parts := strings.Split(input, "-")
		start := net.ParseIP(parts[0])
		end := net.ParseIP(parts[1])

		ips := []string{}
		for ip := start; !ipEqual(ip, end); incIP(ip) {
			ips = append(ips, ip.String())
		}
		ips = append(ips, end.String())
		return ips
	}

	return []string{input}
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] != 0 {
			break
		}
	}
}

func ipEqual(a, b net.IP) bool { return a.String() == b.String() }

func scanPort(ip string, port int, timeout time.Duration) bool {
	addr := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func scanIP(ip string, startPort, endPort int) {
	fmt.Println("Scanning:", ip)

	const workers = 200
	timeout := 20 * time.Millisecond

	jobs := make(chan int, workers)
	results := make(chan int, workers)

	for i := 0; i < workers; i++ {
		go func() {
			for port := range jobs {
				if scanPort(ip, port, timeout) {
					results <- port
				} else {
					results <- 0
				}
			}
		}()
	}

	go func() {
		for port := startPort; port <= endPort; port++ {
			jobs <- port
		}
		close(jobs)
	}()

	done := make(chan bool)

	go func() {
		for port := startPort; port <= endPort; port++ {
			p := <-results
			if p > 0 {
				fmt.Printf("[OPEN] %s:%d\n", ip, p)
			}
		}
		done <- true
	}()

	<-done

	fmt.Println("Finished:", ip)
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: scan.exe <ip/range/cidr> <startPort> <endPort>")
		return
	}

	ipInput := os.Args[1]
	startPort, _ := strconv.Atoi(os.Args[2])
	endPort, _ := strconv.Atoi(os.Args[3])

	ips := parseIPs(ipInput)

	fmt.Println("Scanning started...")

	for _, ip := range ips {
		scanIP(ip, startPort, endPort)
	}

	fmt.Println("Scanning finished")
}
