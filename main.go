package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"example.com/nmap/reporter"
	"example.com/nmap/scanner"
	util "example.com/nmap/utils"
)

func main() {
	cidr := flag.String("cidr", "", "CIDR подсети, например 192.168.1.0/24")
	rangeIP := flag.String("range", "", "Диапазон IP, например 192.168.1.10-192.168.1.20")
	ports := flag.String("ports", "1-1024", "Диапазон портов, например 1-1024 или список 22,80,443")
	out := flag.String("out", "report.json", "JSON файл отчёта")
	html := flag.String("html", "", "Опциональный HTML файл отчёта")
	concurrency := flag.Int("c", 200, "Число concurrent подключений")
	to := flag.Int("timeout", 1500, "Таймаут для подключения и чтения (ms)")
	flag.Parse()

	if *cidr == "" && *rangeIP == "" {
		log.Fatal("Укажите либо -cidr, либо -range")
	}

	// Соберём список IP
	var ips []string
	var err error
	if *cidr != "" {
		ips, err = util.IPsFromCIDR(*cidr)
		if err != nil {
			log.Fatalf("CIDR parse error: %v", err)
		}
	} else {
		ips, err = util.IPsFromRange(*rangeIP)
		if err != nil {
			log.Fatalf("Range parse error: %v", err)
		}
	}

	// Парсим порты
	portList, err := scanner.ParsePorts(*ports)
	if err != nil {
		log.Fatalf("Не удалось распарсить порты: %v", err)
	}

	cfg := scanner.Config{
		TimeoutMS:   *to,
		Concurrency: *concurrency,
	}

	fmt.Printf("Start scanning %d hosts, %d ports each...\n", len(ips), len(portList))
	reports := scanner.ScanHosts(ips, portList, cfg)

	// Сохраним JSON
	err = reporter.SaveJSON(*out, reports)
	if err != nil {
		log.Fatalf("Не удалось сохранить JSON: %v", err)
	}
	fmt.Printf("Saved JSON report: %s\n", *out)

	if strings.TrimSpace(*html) != "" {
		err = reporter.SaveHTML(*html, reports)
		if err != nil {
			log.Fatalf("Не удалось сохранить HTML: %v", err)
		}
		fmt.Printf("Saved HTML report: %s\n", *html)
	}

	fmt.Println("Done")
}
