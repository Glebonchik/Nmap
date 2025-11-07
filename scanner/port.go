package scanner

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

// checkPort пытается подключиться к ip:port и прочитать баннер
func checkPort(ip string, port int, timeoutMS int) PortResult {
	addr := net.JoinHostPort(ip, fmt.Sprintf("%d", port))
	res := PortResult{Port: port, Open: false}
	to := time.Duration(timeoutMS) * time.Millisecond
	conn, err := net.DialTimeout("tcp", addr, to)
	if err != nil {
		// closed/filtered
		return res
	}
	defer conn.Close()
	res.Open = true

	// Установим дедлайн для чтения
	_ = conn.SetReadDeadline(time.Now().Add(to))

	// Для известных HTTP-портов отправим простой запрос, чтобы получить Server header
	if port == 80 || port == 8080 || port == 8000 || port == 443 {
		// Отправляем минимальный HTTP/1.0 запрос
		fmt.Fprintf(conn, "HEAD / HTTP/1.0\r\nHost: %s\r\n\r\n", ip)
	}

	// Попытка прочитать до 1024 байт
	r := bufio.NewReader(conn)
	b, _ := r.Peek(1024) // Peek безопаснее: если нет данных — вернёт ошибку
	s := string(b)

	// Ищем Server: header или первые N символов
	if idx := strings.Index(strings.ToLower(s), "server:"); idx >= 0 {
		// попробуем прочитать строку содержащую Server:
		line, err := r.ReadString('\n')
		if err == nil {
			res.Banner = strings.TrimSpace(line)
		} else {
			res.Banner = strings.TrimSpace(s)
		}
	} else {
		// если нет Server header — возьмём первые 200 символов баннера (может быть баннер SMTP/FTP и т.д.)
		if len(s) > 0 {
			if len(s) > 200 {
				res.Banner = strings.TrimSpace(s[:200])
			} else {
				res.Banner = strings.TrimSpace(s)
			}
		}
	}

	// Нормализация
	res.Banner = strings.ReplaceAll(res.Banner, "\r", "")
	res.Banner = strings.ReplaceAll(res.Banner, "\n", " ")
	res.Banner = strings.TrimSpace(res.Banner)

	return res
}
