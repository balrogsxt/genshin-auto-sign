package main

import (
	"fmt"
	"github.com/balrogsxt/genshin-auto-sign/app"
	"github.com/balrogsxt/genshin-auto-sign/controller"
	"github.com/balrogsxt/genshin-auto-sign/controller/middleware"
	"github.com/balrogsxt/genshin-auto-sign/helper"
	_ "github.com/balrogsxt/genshin-auto-sign/helper"
	"github.com/balrogsxt/genshin-auto-sign/helper/log"
	"github.com/balrogsxt/genshin-auto-sign/task"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"net/http"
)

const Version = "1.0.2"

var _cron *cron.Cron

func main() {
	_cron = cron.New()
	defer func() {
		_cron.Stop()
		if err := recover(); err != nil {
			log.Info("运行异常: %#v", err)
		}
	}()
	log.Info("CurrentVersion: %s", Version)
	conf := initConfig()
	//载入配置文件
	registerTask()
	gin.SetMode(conf.RunMode)
	r := gin.New()
	registerRouter(r)
	uri := fmt.Sprintf("%s:%d", conf.HttpHost, conf.HttpPort)
	log.Info("服务运行在: %s", uri)
	r.Run(uri)
}
func initConfig() *helper.Config {
	conf := helper.GetConfig()
	app.GetRDB().Del(app.GetCtx(), "remoteApiPool")
	conf.CurlApi = append(conf.CurlApi, "live")
	//填装远程api队列到redis
	app.GetRDB().LPush(app.GetCtx(), "remoteApiPool", conf.CurlApi)
	log.Info("载入远程API完成,可调用的远程API数量: %d ", len(conf.CurlApi))
	return conf
}

//注册任务计划
func registerTask() {
	//每1分钟执行一次
	_cron.AddFunc("0 */1 * * * *", func() {
		task.RunSignTask(false)
	})
	//每天凌晨00:00:10执行一次
	_cron.AddFunc("10 0 0 * * *", func() {
		task.RunSignTask(true)
	})
	log.Info("启动任务计划")
	task.RunSignTask(false)

	_cron.Start()
}
func registerRouter(r *gin.Engine) {
	//注册路由中间件
	r.Use(tryCatch())
	//注册跨域允许
	r.Use(CorsAllow())

	//QQ登录
	r.GET("/login", controller.Login)
	r.GET("/loginVerify", controller.LoginVerify)
	r.GET("/getToken", controller.GetToken)

	auth := r.Group("/", middleware.AuthMiddleware)
	{
		auth.GET("/logout", controller.Logout)
		auth.GET("/info", controller.GetInfo)
		auth.POST("/bind", controller.BindPlayer)
		auth.POST("/bindEmail", controller.BindEmail)
	}

}
func CorsAllow() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
	}
}
func tryCatch() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				ex, flag := r.(app.ApiException)
				if flag {
					c.JSON(ex.Code, gin.H{
						"status": ex.Status,
						"msg":    ex.Msg,
					})
				} else {
					log.Info("%s", r)
					c.JSON(500, gin.H{
						"status": 1,
						"msg":    "未知错误",
					})
				}
			}
		}()
		c.Next()
	}
}
