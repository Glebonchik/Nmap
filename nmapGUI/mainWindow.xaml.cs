using System;
using System.Diagnostics;
using System.IO;
using System.Text;
using System.Threading.Tasks;
using System.Windows;

namespace nmapGUI
{
    public partial class MainWindow : Window
    {
        public MainWindow()
        {
            InitializeComponent();
        }

        private async void ScanButton_Click(object sender, RoutedEventArgs e)
        {
            string ip = IpInput.Text.Trim();
            string startPort = StartPortInput.Text.Trim();
            string endPort = EndPortInput.Text.Trim();

            if (string.IsNullOrEmpty(ip) || string.IsNullOrEmpty(startPort) || string.IsNullOrEmpty(endPort))
            {
                MessageBox.Show("Заполните все поля!", "Ошибка", MessageBoxButton.OK, MessageBoxImage.Error);
                return;
            }

            // путь к scanner.exe
            string exePath = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "go", "scanner.exe");
            string exeDir = Path.GetDirectoryName(exePath)!;
            string resultJsonPath = Path.Combine(exeDir, "results.json");

            if (!File.Exists(exePath))
            {
                MessageBox.Show($"scanner.exe не найден:\n{exePath}", "Ошибка", MessageBoxButton.OK, MessageBoxImage.Error);
                return;
            }

            // окно прогресса
            ProgressWindow progress = new ProgressWindow();
            progress.Show();

            try
            {
                // запускаем scanner.exe асинхронно
                await Task.Run(() => RunScanner(exePath, ip, startPort, endPort));

                progress.Close();

                // ЖДЕМ появления файла results.json (иногда Go пишет с задержкой)
                await WaitForFile(resultJsonPath, timeoutMs: 5000);

                string json;

                if (File.Exists(resultJsonPath))
                {
                    json = File.ReadAllText(resultJsonPath, Encoding.UTF8).Trim();
                }
                else
                {
                    json = "{}";
                }

                // окно результата
                ResultWindow wnd = new ResultWindow(json);
                wnd.ShowDialog();
            }
            catch (Exception ex)
            {
                progress.Close();
                MessageBox.Show("Ошибка: " + ex.Message, "Ошибка", MessageBoxButton.OK, MessageBoxImage.Error);
            }
        }

        private void RunScanner(string exePath, string ip, string startPort, string endPort)
        {
            ProcessStartInfo psi = new ProcessStartInfo
            {
                FileName = exePath,
                Arguments = $"-range {ip} -ports {startPort}-{endPort}",
                RedirectStandardOutput = true,
                RedirectStandardError = true,
                UseShellExecute = false,
                CreateNoWindow = true,
                StandardOutputEncoding = Encoding.UTF8,
                WorkingDirectory = Path.GetDirectoryName(exePath) // ВАЖНО!
            };

            using Process proc = new Process { StartInfo = psi };
            proc.Start();

            // читаем stdout, чтобы не было зависаний
            while (!proc.StandardOutput.EndOfStream)
            {
                proc.StandardOutput.ReadLine();
            }

            proc.WaitForExit();
        }

        private async Task WaitForFile(string path, int timeoutMs)
        {
            int waited = 0;
            while (!File.Exists(path) && waited < timeoutMs)
            {
                await Task.Delay(100);
                waited += 100;
            }
        }

        private void AboutButton_Click(object sender, RoutedEventArgs e)
        {
            AboutWindow about = new AboutWindow();
            about.ShowDialog();
        }
    }
}
