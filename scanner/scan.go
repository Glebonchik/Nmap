package scanner

import (
	"fmt"
	"sync"
)

// Config задаёт параметры сканера
type Config struct {
	TimeoutMS   int
	Concurrency int
}

// PortResult описывает результат проверки порта
type PortResult struct {
	Port   int    `json:"port"`
	Open   bool   `json:"open"`
	Banner string `json:"banner,omitempty"`
}

// HostReport содержит результаты по хосту
type HostReport struct {
	IP    string       `json:"ip"`
	Ports []PortResult `json:"ports"`
}

// ParsePorts парсит строку вида "1-1024" или "22,80,443" в слайс портов
func ParsePorts(s string) ([]int, error) {
	var out []int
	// простая поддержка: диапазон или список через запятую
	if s == "" {
		return nil, fmt.Errorf("empty ports")
	}
	if idx := indexOf(s, '-'); idx >= 0 {
		var start, end int
		_, err := fmt.Sscanf(s, "%d-%d", &start, &end)
		if err != nil {
			return nil, err
		}
		for p := start; p <= end; p++ {
			out = append(out, p)
		}
		return out, nil
	}
	// иначе список
	var p int
	for _, part := range splitAndTrim(s, ',') {
		_, err := fmt.Sscanf(part, "%d", &p)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

func indexOf(s string, ch rune) int {
	for i, r := range s {
		if r == ch {
			return i
		}
	}
	return -1
}

func splitAndTrim(s string, sep rune) []string {
	var parts []string
	cur := ""
	for _, r := range s {
		if r == sep {
			parts = append(parts, trim(cur))
			cur = ""
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		parts = append(parts, trim(cur))
	}
	return parts
}

func trim(s string) string {
	return string([]byte(s))
}

// ScanHosts сканирует список IP и возвращает отчёты
func ScanHosts(ips []string, ports []int, cfg Config) []HostReport {
	var wg sync.WaitGroup
	in := make(chan string)
	out := make(chan HostReport)

	// worker pool
	workers := cfg.Concurrency
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range in {
				rep := scanSingleHost(ip, ports, cfg)
				out <- rep
			}
		}()
	}

	// feeder
	go func() {
		for _, ip := range ips {
			in <- ip
		}
		close(in)
	}()

	// collector
	var reports []HostReport
	go func() {
		wg.Wait()
		close(out)
	}()

	for r := range out {
		reports = append(reports, r)
	}

	return reports
}

// scanSingleHost сканирует порты у одного хоста
func scanSingleHost(ip string, ports []int, cfg Config) HostReport {
	var res HostReport
	res.IP = ip
	for _, p := range ports {
		pr := checkPort(ip, p, cfg.TimeoutMS)
		res.Ports = append(res.Ports, pr)
		// небольшая оптимизация: если много портов — можно распараллелить внутри
	}
	return res
}
