package main

import (
	"fmt"
	"github.com/balrogsxt/genshin-auto-sign/app"
	"github.com/balrogsxt/genshin-auto-sign/controller"
	"github.com/balrogsxt/genshin-auto-sign/controller/middleware"
	"github.com/balrogsxt/genshin-auto-sign/helper"
	_ "github.com/balrogsxt/genshin-auto-sign/helper"
	"github.com/balrogsxt/genshin-auto-sign/task"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
	"net/http"
)

const Version = "1.0.0"

var _cron *cron.Cron

func main() {
	_cron = cron.New()
	defer func() {
		_cron.Stop()
		if err := recover(); err != nil {
			fmt.Printf("运行异常: %#v \n", err)
		}
	}()
	fmt.Println("CurrentVersion: ", Version)
	conf := helper.GetConfig()

	//载入配置文件
	registerTask()
	gin.SetMode(conf.RunMode)
	r := gin.New()
	registerRouter(r)
	uri := fmt.Sprintf("%s:%d", conf.HttpHost, conf.HttpPort)
	fmt.Println("服务运行在: ", uri)
	r.Run(uri)
}

//注册任务计划
func registerTask() {
	//每5分钟执行一次
	_cron.AddFunc("0 */5 * * * *", func() {
		task.RunSignTask(false)
	})
	//每天凌晨00:00:10执行一次
	_cron.AddFunc("10 0 0 * * *", func() {
		task.RunSignTask(true)
	})

	fmt.Println("启动任务计划")
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
					fmt.Printf("%s \n", r)
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
