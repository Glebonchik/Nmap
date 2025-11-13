package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"example.com/nmap/api"
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
	apiFlag := flag.String("api", "", "start HTTP API (e.g. :8080)")
	flag.Parse()

	// Если указан -api, запускаем API-сервер в фоновом режиме.
	if strings.TrimSpace(*apiFlag) != "" {
		go api.StartAPIServer(*apiFlag)
		log.Printf("API server started on %s", *apiFlag)
	}

	// Если не указаны ни -cidr, ни -range, и -api был указан — просто заблокироваться,
	// чтобы API оставался доступен (это режим 'API-only').
	if *cidr == "" && *rangeIP == "" {
		if strings.TrimSpace(*apiFlag) != "" {
			// блокируем главный поток, API работает в горутине
			select {}
		} else {
			log.Fatal("Укажите либо -cidr, либо -range (или запустите с -api для режима сервера)")
		}
	}

	// Если тут — значит у нас есть цель для сканирования
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

	// Если сервер API был запущен, не завершаем сразу: даём ему время поработать.
	if strings.TrimSpace(*apiFlag) != "" {
		fmt.Println("Scan finished. API server still running. Press Ctrl+C to exit.")
		// просто блокируем главный поток, чтобы процесс не завершался немедленно
		select {}
	}

	fmt.Println("Done")
}
