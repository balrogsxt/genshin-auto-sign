package helper

import (
	"fmt"
	"github.com/balrogsxt/genshin-auto-sign/helper/log"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var conf *Config = nil

func init() {
	conf = LoadConfig()
}

type Config struct {
	RunMode          string   `yaml:"run_mode"`           //运行模式release=生产环境 test=测试环境 debug=调试环境
	RedirectTokenUrl string   `yaml:"redirect_token_url"` //登录成功后的回调token地址 ,变量%token%
	HttpHost         string   `yaml:"http_host"`          //服务绑定地址
	HttpPort         int      `yaml:"http_port"`          //服务启动端口
	JwtKey           string   `yaml:"jwt_key"`            //jwt密钥
	NewUser          bool     `yaml:"new_user"`           //是否允许新用户使用
	CurlApi          []string `yaml:"curl_api"`           //远程请求API
	Task             []string `yaml:"task"`               //任务corn触发时间
	QQOauth          struct {
		ClientId     string `yaml:"client_id"`
		ClientSecret string `yaml:"client_secret"`
		RedirectUri  string `yaml:"redirect_uri"`
	}
	Mysql struct {
		Host     string
		Port     int
		Name     string
		User     string
		Password string
	}
	Redis struct {
		Host     string
		Port     int
		Password string
		Index    int
	}
	QQBot struct {
		Url               string
		QQ                string
		Key               string
		BindNotifyGroup   []string `yaml:"bind_notify_group"`   //绑定用户成功后通知的群组
		SignNotifyGroup   []string `yaml:"sign_notify_group"`   //签到成功后通知的群组
		ExpireNotifyGroup []string `yaml:"expire_notify_group"` //cookie过期后通知的群组
	}
}

//获取配置文件
func GetConfig() *Config {
	if conf == nil {
		conf = LoadConfig()
	}
	return conf
}

//加载配置文件
func LoadConfig() *Config {
	file, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		panic(fmt.Sprintf("读取配置文件失败: %s", err.Error()))
	}
	conf := Config{}
	if err := yaml.Unmarshal(file, &conf); err != nil {
		panic(fmt.Sprintf("解析配置文件失败: %s", err.Error()))
	}
	log.Info("载入配置文件成功")
	return &conf
}
