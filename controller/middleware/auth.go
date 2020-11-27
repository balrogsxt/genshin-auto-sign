package middleware

import (
	"fmt"
	"github.com/balrogsxt/genshin-auto-sign/app"
	"github.com/balrogsxt/genshin-auto-sign/helper"
	"github.com/gin-gonic/gin"
	"strconv"
)

//验证是否登录中间件
func AuthMiddleware(c *gin.Context) {
	Authorization := c.GetHeader("Authorization")
	if helper.IsEmpty(Authorization) {
		c.Abort()
		app.NewException("登录失效,请重新登录", 403)
	}
	_map, err := helper.JwtParse(Authorization)
	if err != nil {
		c.Abort()
		app.NewException("登录失效,请重新登录", 403)
	}

	//判断是否过期
	md5Token := helper.Md5(Authorization)
	if app.GetRDB().Exists(app.GetCtx(), fmt.Sprintf("genshinToken:%s", md5Token)).Val() == 0 {
		c.Abort()
		app.NewException("登录已过期,请重新登录", 403)
	}

	userId, _ := strconv.ParseInt(fmt.Sprintf("%v", _map["userId"]), 10, 64)
	//验证用户是否存在
	um := app.UserModel{}
	has, _ := app.GetDb().Where("id = ?", userId).Exist(&um)
	if !has {
		//用户不存在?
		c.Abort()
		app.NewException("用户失效,请重新登录", 403)
	}

	//设置数据传递
	c.Set("userid", userId)
	c.Set("openid", _map["openid"])
}
