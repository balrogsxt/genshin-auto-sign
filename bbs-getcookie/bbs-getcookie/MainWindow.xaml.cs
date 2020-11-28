using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Data;
using System.Windows.Documents;
using System.Windows.Input;
using System.Windows.Media;
using System.Windows.Media.Imaging;
using System.Windows.Navigation;
using mshtml;

namespace bbs_getcookie {
    /// <summary>
    /// MainWindow.xaml 的交互逻辑
    /// </summary>
    public partial class MainWindow : Window {
        public MainWindow() {
            InitializeComponent();

            this.Loaded += MainWindow_Loaded;
        
        }

        private void MainWindow_Loaded(object sender, RoutedEventArgs ev) {
            this.web.Navigate(new Uri("https://user.mihoyo.com/?cb_url=https%3A%2F%2Fbbs.mihoyo.com%2Fys%2F&week=1#/login/password"));

            var doc = this.web.Document as HTMLDocument;

            this.getCookie.Click += (s, eve) => {
                try {
                    var aid = string.Empty;
                    var ct = string.Empty;

                    var cookie = doc.cookie;
                    var list = cookie.Trim().Split(';');
                    foreach (var item in list) {
                        var it = item.Split('=');
                        if(it.Length >= 2) {
                            var name = it[0].Trim().ToLower();
                            if (name.Equals("account_id")) {
                                aid = it[1];
                            } else if (name.Equals("cookie_token")) {
                                ct = it[1];
                            }
                        }
                    }
                    if (aid.Length == 0 || ct.Length == 0) {
                        MessageBox.Show("可能未登录,请先登录!");
                    } else {
                        this.account_id.Text = aid;
                        this.cookie_token.Text = ct;
                        this.data.Visibility = Visibility.Visible;
                        MessageBox.Show("获取成功,请在右边获取数据并复制到https://genshin.acgxt.com中填写绑定!");

                    }
                } catch(Exception e) {
                    MessageBox.Show("发生错误:"+e.Message);
                }
            };
        }
    }
}
