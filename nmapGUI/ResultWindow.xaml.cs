using System.Windows;

namespace nmapGUI
{
    public partial class ResultWindow : Window
    {
        public ResultWindow(string jsonResult)
        {
            InitializeComponent();
            JsonTextBox.Text = jsonResult;
        }
    }
}
