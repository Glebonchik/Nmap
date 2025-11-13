package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"example.com/nmap/scanner"
	util "example.com/nmap/utils"
)

// Simple controller state
var (
	mu        sync.Mutex
	running   bool
	startedAt time.Time
	doneAt    time.Time
	lastErr   string
	result    []scanner.HostReport
)

// ScanRequest — структура POST /scan
type ScanRequest struct {
	CIDR        string `json:"cidr,omitempty"`
	Range       string `json:"range,omitempty"`
	Ports       string `json:"ports"` // e.g. "1-1024" or "22,80,443"
	Concurrency int    `json:"concurrency"`
	TimeoutMS   int    `json:"timeout_ms"`
}

// StatusResponse — статус выполнения
type StatusResponse struct {
	Running   bool   `json:"running"`
	StartedAt string `json:"started_at,omitempty"`
	DoneAt    string `json:"done_at,omitempty"`
	LastError string `json:"last_error,omitempty"`
	Hosts     int    `json:"hosts,omitempty"`
}

// StartScanHandler - запускаем сканирование (non-blocking)
func StartScanHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only POST", http.StatusMethodNotAllowed)
		return
	}

	var req ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	if running {
		mu.Unlock()
		http.Error(w, "scan already running", http.StatusConflict)
		return
	}
	// mark running
	running = true
	startedAt = time.Now()
	doneAt = time.Time{}
	lastErr = ""
	result = nil
	mu.Unlock()

	// run scan asynchronously
	go func(rq ScanRequest) {
		defer func() {
			mu.Lock()
			running = false
			doneAt = time.Now()
			mu.Unlock()
		}()

		// build IP list
		var ips []string
		var err error
		if rq.CIDR != "" {
			ips, err = util.IPsFromCIDR(rq.CIDR)
		} else if rq.Range != "" {
			ips, err = util.IPsFromRange(rq.Range)
		} else {
			err = &json.UnmarshalTypeError{} // custom error: no target
		}
		if err != nil {
			mu.Lock()
			lastErr = "ip parse error: " + err.Error()
			mu.Unlock()
			return
		}

		ports, err := scanner.ParsePorts(rq.Ports)
		if err != nil {
			mu.Lock()
			lastErr = "ports parse error: " + err.Error()
			mu.Unlock()
			return
		}

		cfg := scanner.Config{
			TimeoutMS:   rq.TimeoutMS,
			Concurrency: rq.Concurrency,
		}
		log.Printf("API: starting scan: hosts=%d ports=%d", len(ips), len(ports))
		out := scanner.ScanHosts(ips, ports, cfg)

		mu.Lock()
		result = out
		mu.Unlock()
	}(req)

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"started"}`))
}

// StatusHandler - возвращает статус
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	resp := StatusResponse{
		Running:   running,
		LastError: lastErr,
	}
	if !startedAt.IsZero() {
		resp.StartedAt = startedAt.Format(time.RFC3339)
	}
	if !doneAt.IsZero() {
		resp.DoneAt = doneAt.Format(time.RFC3339)
	}
	if result != nil {
		resp.Hosts = len(result)
	}
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ReportHandler - возвращает JSON-отчёт
func ReportHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	out := result
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if out == nil {
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte(`[]`))
		return
	}
	json.NewEncoder(w).Encode(out)
}

// StartAPIServer запускает HTTP-сервер на указанном адресе
func StartAPIServer(addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/scan", StartScanHandler)
	mux.HandleFunc("/status", StatusHandler)
	mux.HandleFunc("/report", ReportHandler)

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	log.Printf("API server listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("API server failed: %v", err)
	}
}
