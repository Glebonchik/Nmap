using System;
using System.Diagnostics;
using System.IO;
using System.Text;
using System.Windows;

namespace nmapGUI
{
    public partial class MainWindow : Window
    {
        public MainWindow()
        {
            InitializeComponent();
        }

        private void ScanButton_Click(object sender, RoutedEventArgs e)
        {
            string ip = IpTextBox.Text.Trim();
            string ports = PortsTextBox.Text.Trim();

            if (string.IsNullOrEmpty(ip))
            {
                MessageBox.Show("Введите IP/CIDR/Range.", "Ошибка", MessageBoxButton.OK, MessageBoxImage.Warning);
                return;
            }

            // Путь к Go exe
            string goExe = Path.Combine(AppDomain.CurrentDomain.BaseDirectory,"go","scanner.exe");

            if (!File.Exists(goExe))
            {
                MessageBox.Show("Файл scanner.exe не найден!", "Ошибка", MessageBoxButton.OK, MessageBoxImage.Error);
                return;
            }

            // Формируем аргументы
            string args = $"-range {ip} -ports {ports} -out result.json";

            // Запускаем процесс
            ProcessStartInfo psi = new ProcessStartInfo
            {
                FileName = goExe,
                Arguments = args,
                RedirectStandardOutput = true,
                RedirectStandardError = true,
                UseShellExecute = false,
                CreateNoWindow = true,
                StandardOutputEncoding = Encoding.UTF8,
                StandardErrorEncoding = Encoding.UTF8
            };

            try
            {
                Process proc = new Process { StartInfo = psi };
                proc.Start();

                StringBuilder output = new StringBuilder();

                // Читаем stdout
                while (!proc.StandardOutput.EndOfStream)
                {
                    string line = proc.StandardOutput.ReadLine();
                    output.AppendLine(line);
                }

                // Читаем stderr
                while (!proc.StandardError.EndOfStream)
                {
                    string line = proc.StandardError.ReadLine();
                    output.AppendLine(line);
                }

                proc.WaitForExit();

                // Показываем результаты в новом окне
                ResultWindow resultWindow = new ResultWindow(output.ToString());
                resultWindow.Show();
            }
            catch (Exception ex)
            {
                MessageBox.Show($"Ошибка при запуске сканера: {ex.Message}", "Ошибка", MessageBoxButton.OK, MessageBoxImage.Error);
            }
        }

        private void AboutButton_Click(object sender, RoutedEventArgs e)
        {
            string aboutText = "Mini-Nmap GUI\nАвтор: Трушин Г.А.\nКафедра: МАИ 316\n\nИнструкция:\n1. Введите IP/CIDR/Range.\n2. Укажите порты.\n3. Нажмите 'Сканировать'.\n4. Результаты откроются в новом окне.";
            MessageBox.Show(aboutText, "О программе", MessageBoxButton.OK, MessageBoxImage.Information);
        }
    }
}
