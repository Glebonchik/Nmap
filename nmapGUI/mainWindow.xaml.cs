// MainWindow.xaml.cs
using System;
using System.Collections.Generic;
using System.Net.Http;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Threading;

namespace MiniNmapGUI
{
    public partial class MainWindow : Window
    {
        private readonly HttpClient _http = new HttpClient() { BaseAddress = new Uri("http://localhost:8080") };
        private DispatcherTimer _timer;

        public MainWindow()
        {
            InitializeComponent();

            // простой таймер, который опрашивает статус каждые 2 секунды
            _timer = new DispatcherTimer();
            _timer.Interval = TimeSpan.FromSeconds(2);
            _timer.Tick += Timer_Tick;
        }

        private async void StartBtn_Click(object sender, RoutedEventArgs e)
        {
            StartBtn.IsEnabled = false;
            StatusLabel.Content = "Starting...";
            ProgressBar.Value = 0;

            var req = new
            {
                cidr = string.IsNullOrWhiteSpace(CidrBox.Text) ? null : CidrBox.Text,
                range = string.IsNullOrWhiteSpace(RangeBox.Text) ? null : RangeBox.Text,
                ports = PortsBox.Text,
                concurrency = int.TryParse(ConcBox.Text, out var c) ? c : 200,
                timeout_ms = int.TryParse(ToBox.Text, out var t) ? t : 1500
            };

            var json = JsonSerializer.Serialize(req);
            var content = new StringContent(json, Encoding.UTF8, "application/json");
            try
            {
                var resp = await _http.PostAsync("/scan", content);
                if (resp.IsSuccessStatusCode || resp.StatusCode == System.Net.HttpStatusCode.Accepted)
                {
                    StatusLabel.Content = "Scan started";
                    _timer.Start();
                }
                else
                {
                    var text = await resp.Content.ReadAsStringAsync();
                    StatusLabel.Content = $"Error: {resp.StatusCode} {text}";
                    StartBtn.IsEnabled = true;
                }
            }
            catch (Exception ex)
            {
                StatusLabel.Content = "HTTP error: " + ex.Message;
                StartBtn.IsEnabled = true;
            }
        }

        private async void Timer_Tick(object sender, EventArgs e)
        {
            try
            {
                var resp = await _http.GetAsync("/status");
                if (!resp.IsSuccessStatusCode)
                {
                    StatusLabel.Content = "Status error";
                    return;
                }
                var txt = await resp.Content.ReadAsStringAsync();
                var status = JsonSerializer.Deserialize<StatusResponse>(txt, new JsonSerializerOptions { PropertyNameCaseInsensitive = true });

                if (status == null) return;

                StatusLabel.Content = status.Running ? $"Running (hosts={status.Hosts})" : $"Idle (hosts={status.Hosts})";
                // Простое эмпирическое движение прогресса
                ProgressBar.IsIndeterminate = status.Running;
                if (!status.Running)
                {
                    _timer.Stop();
                    StartBtn.IsEnabled = true;
                }
            }
            catch (Exception ex)
            {
                StatusLabel.Content = "Status error: " + ex.Message;
            }
        }

        private async void GetReportBtn_Click(object sender, RoutedEventArgs e)
        {
            try
            {
                var resp = await _http.GetAsync("/report");
                if (!resp.IsSuccessStatusCode)
                {
                    MessageBox.Show("No report or error: " + resp.StatusCode);
                    return;
                }
                var txt = await resp.Content.ReadAsStringAsync();
                var reports = JsonSerializer.Deserialize<List<HostReport>>(txt, new JsonSerializerOptions { PropertyNameCaseInsensitive = true });

                // Show summary in DataGrid
                var summary = new List<HostSummary>();
                foreach (var h in reports)
                {
                    int openCount = 0;
                    foreach (var p in h.Ports) if (p.Open) openCount++;
                    summary.Add(new HostSummary { IP = h.IP, OpenCount = openCount });
                }
                ResultsGrid.ItemsSource = summary;

                // Save JSON to temp file and also create a simple HTML to view in WebBrowser
                var tmpJson = System.IO.Path.Combine(System.IO.Path.GetTempPath(), "mini-nmap-report.json");
                await System.IO.File.WriteAllTextAsync(tmpJson, txt);

                // Optionally, fetch HTML from separate endpoint; for now render a basic HTML
                var html = BuildHtml(reports);
                var tmpHtml = System.IO.Path.Combine(System.IO.Path.GetTempPath(), "mini-nmap-report.html");
                await System.IO.File.WriteAllTextAsync(tmpHtml, html, Encoding.UTF8);

                Browser.Navigate(new Uri(tmpHtml));
            }
            catch (Exception ex)
            {
                MessageBox.Show("Error fetching report: " + ex.Message);
            }
        }

        private string BuildHtml(List<HostReport> reports)
        {
            var sb = new StringBuilder();
            sb.AppendLine("<!doctype html><html><head><meta charset='utf-8'><title>MiniNmap Report</title></head><body>");
            sb.AppendLine("<h1>Report</h1>");
            sb.AppendLine($"<p>Generated: {DateTime.Now}</p>");
            foreach (var h in reports)
            {
                sb.AppendLine($"<h2>Host: {h.IP}</h2>");
                sb.AppendLine("<table border='1' cellpadding='4' cellspacing='0'><tr><th>Port</th><th>Open</th><th>Banner</th></tr>");
                foreach (var p in h.Ports)
                {
                    sb.AppendLine($"<tr><td>{p.Port}</td><td>{p.Open}</td><td>{System.Net.WebUtility.HtmlEncode(p.Banner)}</td></tr>");
                }
                sb.AppendLine("</table>");
            }
            sb.AppendLine("</body></html>");
            return sb.ToString();
        }
    }

    // helper models matching Go structures
    public class HostReport
    {
        public string IP { get; set; }
        public List<PortResult> Ports { get; set; }
    }
    public class PortResult
    {
        public int Port { get; set; }
        public bool Open { get; set; }
        public string Banner { get; set; }
    }
    public class StatusResponse
    {
        public bool Running { get; set; }
        public string StartedAt { get; set; }
        public string DoneAt { get; set; }
        public string LastError { get; set; }
        public int Hosts { get; set; }
    }
    public class HostSummary
    {
        public string IP { get; set; }
        public int OpenCount { get; set; }
    }
}
