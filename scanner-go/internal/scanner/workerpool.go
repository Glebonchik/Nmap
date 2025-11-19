package scanner

import (
	"time"
)

type Task struct {
	IP string
}

// Сканирует список IP параллельно
func ScanAllHosts(ips []string, ports []int, timeout time.Duration, workers int) []HostScan {
	tasks := make(chan Task, len(ips))
	results := make(chan HostScan, len(ips))

	// Стартуем воркеров
	for i := 0; i < workers; i++ {
		go worker(tasks, results, ports, timeout)
	}

	// Отправляем задания
	for _, ip := range ips {
		tasks <- Task{IP: ip}
	}
	close(tasks)

	// Собираем результаты
	scans := make([]HostScan, 0, len(ips))
	for i := 0; i < len(ips); i++ {
		scans = append(scans, <-results)
	}

	return scans
}

// Воркер, обрабатывает IP-адреса
func worker(tasks <-chan Task, results chan<- HostScan, ports []int, timeout time.Duration) {
	for task := range tasks {
		hostResult := ScanPorts(task.IP, ports, timeout)
		results <- hostResult
	}
}
