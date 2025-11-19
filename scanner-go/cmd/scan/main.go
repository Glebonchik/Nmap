package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"scanner-go/internal/ip"
	"scanner-go/internal/report"
	"scanner-go/internal/scanner"
)

func main() {

	// Флаги для CLI
	ipRange := flag.String("range", "", "Диапазон IP (например 192.168.1.1-192.168.1.20 или CIDR 192.168.1.0/24)")
	portsArg := flag.String("ports", "1-1024", "Порты (например 1-1024 или 22,80,443)")
	outFile := flag.String("out", "results.json", "Файл для сохранения результата JSON")
	workers := flag.Int("workers", 100, "Количество параллельных воркеров")
	timeout := flag.Int("timeout", 20, "Таймаут для каждого порта (ms)")

	flag.Parse()

	if *ipRange == "" {
		log.Fatal("Укажите диапазон IP через -range")
	}

	// Парсинг
	ips, err := ip.ParseAny(*ipRange)
	if err != nil {
		log.Fatal("Ошибка в диапазоне IP: ", err)
	}

	ports, err := parsePorts(*portsArg)
	if err != nil {
		log.Fatal("Ошибка в портах: ", err)
	}

	// Параметры
	scanTimeout := time.Duration(*timeout) * time.Millisecond

	fmt.Println("Начало сканирования...")
	// Сканирование
	results := scanner.ScanAllHosts(ips, ports, scanTimeout, *workers)

	// Сохранение отчета
	err = report.SaveJSON(*outFile, results)
	if err != nil {
		log.Fatal("Ошибка при сохранении JSON: ", err)
	}

	fmt.Println("Сканичрование завершено!")
	fmt.Println("Сохранено в:", *outFile)
}

// Превращает ввод пользователя в массив
func parsePorts(arg string) ([]int, error) {
	if strings.Contains(arg, "-") {
		parts := strings.Split(arg, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("Неверный формат диапазона портов")
		}

		start, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}
		end, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}
		if start > end {
			return nil, fmt.Errorf("Диапазон указан неверно")
		}

		res := make([]int, 0)
		for p := start; p <= end; p++ {
			res = append(res, p)
		}
		return res, nil
	}

	// Список портов
	items := strings.Split(arg, ",")
	res := make([]int, 0, len(items))

	for _, s := range items {
		p, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil {
			return nil, err
		}
		res = append(res, p)
	}

	return res, nil
}
