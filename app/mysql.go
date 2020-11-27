package app

import (
	"fmt"
	"github.com/balrogsxt/genshin-auto-sign/helper"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
)

var engine *xorm.Engine

func init() {
	conf := helper.GetConfig().Mysql
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8", conf.User, conf.Password, conf.Host, conf.Port, conf.Name)
	engine, _ = xorm.NewEngine("mysql", dsn)
	err := engine.Ping()
	if err != nil {
		fmt.Println("mysql连接数据库失败")
	} else {
		fmt.Println("mysql连接数据库成功!!!")
	}
}
func GetDb() *xorm.Engine {
	return engine
}

type UserModel struct {
	Id              int    `xorm:"pk autoincr"` //自增ID
	OpenId          string `xorm:"openid"`      //qqopenid
	MihoyoAccountId string `xorm:"account_id"`  //米游社账户Id
	MihoyoWebToken  string `xorm:"web_token"`   //米游社账户Token
	CreateTime      int64  `xorm:"create_time"` //注册时间
	PlayerUid       string `xorm:"player_id"`   //玩家Uid
	PlayerName      string `xorm:"player_name"` //玩家名称
	ServerName      string `xorm:"server_name"` //服务器名称
	BindTime        int64  `xorm:"bind_time"`   //绑定时间
	SignTime        int64  `xorm:"sign_time"`   //签到时间
	TotalSign       int    `xorm:"total_sign"`  //累计签到天数
}

func (t *UserModel) TableName() string {
	return "user"
}
