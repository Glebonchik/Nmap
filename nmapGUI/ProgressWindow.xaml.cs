using System.Windows;

namespace nmapGUI
{
    public partial class ProgressWindow : Window
    {
        public ProgressWindow()
        {
            InitializeComponent();
        }

        /// <summary>
        /// Если нужно, можно добавить метод для обновления прогресса,
        /// но сейчас прогресс индетерминированный (ползунок двигается сам)
        /// </summary>
        /// <param name="percent">0-100</param>
        public void UpdateProgress(int percent)
        {
            // Пока что прогресс просто визуальный, без привязки к %.
            // Можно реализовать через Dispatcher.Invoke и ProgressBar.Value
        }
    }
}
