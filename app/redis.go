package app

import (
	"context"
	"fmt"
	"github.com/balrogsxt/genshin-auto-sign/helper"
	"github.com/balrogsxt/genshin-auto-sign/helper/log"
	"github.com/go-redis/redis/v8"
)

//这里初始化redis数据库连接
var ctx = context.Background()
var rdb *redis.Client

func init() {
	initRDB()
}
func initRDB() {
	conf := helper.GetConfig().Redis
	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", conf.Host, conf.Port),
		DB:       conf.Index,
		Password: conf.Password,
	})
	err := rdb.Ping(ctx).Err()
	if err != nil {
		log.Info("Redis连接异常!")
	} else {
		log.Info("Redis已连接成功!!!!")
	}
}
func GetCtx() context.Context {
	return ctx
}

//获取redis连接
func GetRDB() *redis.Client {
	if rdb == nil {
		initRDB()
	}
	return rdb
}
