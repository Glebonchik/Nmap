using System.Windows;
using System.Windows.Threading;

namespace nmapGUI
{
    public partial class ProgressWindow : Window
    {
        private DispatcherTimer timer;
        private int progress = 0;

        public ProgressWindow()
        {
            InitializeComponent();

            timer = new DispatcherTimer();
            timer.Interval = TimeSpan.FromMilliseconds(150);
            timer.Tick += Timer_Tick;
            timer.Start();
        }

        private void Timer_Tick(object? sender, EventArgs e)
        {
            progress += 2;
            if (progress >= 100)
                progress = 10;

            ScanProgressBar.Value = progress;
        }

        public void StopProgress()
        {
            timer?.Stop();
        }
    }
}
