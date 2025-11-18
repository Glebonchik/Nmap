package reporter

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"

	"nmap/scanner"
)

// -----------------------------
// Сохранение JSON отчёта
// -----------------------------
func SaveJSON(filename string, data []scanner.HostScan) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(data)
}

// -----------------------------
// Сохранение HTML отчёта
// -----------------------------
const htmlTemplate = `
<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="UTF-8"/>
	<title>Мини Nmap — отчет</title>
	<style>
		body { font-family: Arial, sans-serif; background: #f2f2f2; padding: 20px; }
		h1 { color: #333; }
		table { width: 100%; border-collapse: collapse; margin-top: 20px; }
		th, td { border: 1px solid #ccc; padding: 8px; text-align: left; }
		th { background: #555; color: white; }
		tr:nth-child(even) { background: #eee; }
		.ok { color: green; font-weight: bold; }
		.bad { color: red; font-weight: bold; }
	</style>
</head>
<body>

<h1>Отчёт о сканировании сети</h1>

{{range .}}
	<h2>Хост: {{.IP}}</h2>
	<table>
		<tr>
			<th>Порт</th>
			<th>Открыт</th>
			<th>TCP-баннер</th>
			<th>TLS версия</th>
			<th>TLS шифр</th>
		</tr>

		{{range .Ports}}
		<tr>
			<td>{{.Port}}</td>
			<td>
				{{if .Open}}
					<span class="ok">Да</span>
				{{else}}
					<span class="bad">Нет</span>
				{{end}}
			</td>
			<td>{{.Banner}}</td>
			<td>{{.TLSVersion}}</td>
			<td>{{.TLSCipher}}</td>
		</tr>
		{{end}}
	</table>
{{end}}

</body>
</html>
`

func SaveHTML(filename string, data []scanner.HostScan) error {
	tmpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, data)
}

// -----------------------------
// Унифицированный сохранитель
// -----------------------------
func SaveReport(jsonFile string, htmlFile string, data []scanner.HostScan) error {
	if err := SaveJSON(jsonFile, data); err != nil {
		return fmt.Errorf("json error: %v", err)
	}
	if htmlFile != "" {
		if err := SaveHTML(htmlFile, data); err != nil {
			return fmt.Errorf("html error: %v", err)
		}
	}
	return nil
}
