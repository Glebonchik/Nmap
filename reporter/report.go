package reporter

import (
	"encoding/json"
	"os"
	"text/template"
	"time"

	"example.com/nmap/scanner"
)

func SaveJSON(path string, reports []scanner.HostReport) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", " ")
	return enc.Encode(reports)
}

// SaveHTML создаёт простой HTML-отчёт
func SaveHTML(path string, reports []scanner.HostReport) error {
	tpl := `<!doctype html>
<html>
<head>
<meta charset="utf-8">
<title>Mini Nmap Report</title>
<style>
body{font-family: Arial, sans-serif;}
table{border-collapse: collapse;width:100%;}
th,td{border:1px solid #ddd;padding:8px}
th{background:#f4f4f4}
</style>
</head>
<body>
<h1>Mini Nmap Report</h1>
<p>Generated: {{.Now}}</p>
{{range .Reports}}
<h2>Host: {{.IP}}</h2>
<table>
<thead><tr><th>Port</th><th>Open</th><th>Banner</th></tr></thead>
<tbody>
{{range .Ports}}
<tr><td>{{.Port}}</td><td>{{.Open}}</td><td>{{.Banner}}</td></tr>
{{end}}
</tbody>
</table>
{{end}}
</body>
</html>`

	t := template.New("report")
	t, err := t.Parse(tpl)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	data := struct {
		Now     string
		Reports []scanner.HostReport
	}{
		Now:     time.Now().Format(time.RFC1123),
		Reports: reports,
	}

	return t.Execute(f, data)
}
