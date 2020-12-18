package app

import (
	"fmt"
	"github.com/balrogsxt/genshin-auto-sign/helper"
	"github.com/balrogsxt/genshin-auto-sign/helper/log"
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
		log.Info("mysql连接数据库失败")
	} else {
		log.Info("mysql连接数据库成功!!!")
	}
}
func GetDb() *xorm.Engine {
	return engine
}
