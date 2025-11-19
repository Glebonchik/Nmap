package report

import (
	"encoding/json"
	"os"

	"scanner-go/internal/scanner"
)

// Сохраняет результаты сканирования в файл
func SaveJSON(filename string, results []scanner.HostScan) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}
