using System.Windows;

namespace nmapGUI
{
    public partial class ResultWindow : Window
    {
        public ResultWindow(string log)
        {
            InitializeComponent();
            ResultTextBox.Text = log;
        }
    }
}
