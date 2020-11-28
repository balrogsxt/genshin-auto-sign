package controller

import (
	"fmt"
	"github.com/balrogsxt/genshin-auto-sign/api"
	"github.com/balrogsxt/genshin-auto-sign/app"
	"github.com/balrogsxt/genshin-auto-sign/helper"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"math/rand"
	"strings"
	"time"
)

//登录QQ
func Login(c *gin.Context) {
	qq := api.NewQQ()
	url := qq.BuildLoginUrl()
	c.Redirect(302, url)
}

//退出登录
func Logout(c *gin.Context) {
	token := c.GetHeader("authorization")
	md5Token := helper.Md5(token)
	//删除
	app.GetRDB().Del(app.GetCtx(), fmt.Sprintf("genshinToken:%s", md5Token))
	c.JSON(200, gin.H{
		"status": 0,
		"msg":    "退出登录成功",
	})
}

//qq回调验证登录
func LoginVerify(c *gin.Context) {
	code := c.Query("code")
	qq := api.NewQQ()
	_accessToken, err := qq.GetAccessToken(code)
	if err != nil {
		app.NewException("登录失败:" + err.Error())
	}
	_openid, err := qq.GetOpenId(_accessToken.AccessToken)
	if err != nil {
		app.NewException("登录失败:" + err.Error())
	}
	openid := _openid.OpenId

	um := app.UserModel{}
	var userId int
	//验证是否注册过账户
	has, err := app.GetDb().Where("openid = ?", openid).Get(&um)
	if has == false {
		//账户未注册,先注册
		userModel := new(app.UserModel)
		userModel.OpenId = openid
		userModel.CreateTime = time.Now().Unix()
		_, err := app.GetDb().Insert(userModel)
		if err != nil {
			fmt.Println("注册失败:", err.Error())
			app.NewException("账户注册失败...请稍后再试!")
		}
		userId = userModel.Id
	} else {
		userId = um.Id
	}
	//判断是否注册
	jwtToken, err := helper.JwtBuild(jwt.MapClaims{
		"userId": userId,
		"openid": openid,
		"at":     time.Now().Unix(), //创建时间
	})
	if err != nil {
		app.NewException("账户授权失败...请稍后再试!")
	}
	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(9999999)
	tmpToken := helper.Md5(fmt.Sprintf("%s_%d", jwtToken, r))
	md5Token := helper.Md5(jwtToken)
	//获取jwt的临时token,有效期60秒,获取删除
	app.GetRDB().SetEX(app.GetCtx(), fmt.Sprintf("genshin_tmp_token:%s", tmpToken), jwtToken, time.Second*60)
	//token真实有效期,7天内有效,每次访问增加时间
	app.GetRDB().SetEX(app.GetCtx(), fmt.Sprintf("genshinToken:%s", md5Token), 1, time.Second*86400*7)

	//带着数据重定向过去
	conf := helper.GetConfig()
	redirect := strings.ReplaceAll(conf.RedirectTokenUrl, "%token%", tmpToken)
	c.Redirect(302, redirect)
}

//获取当前登录用户信息状态
func GetInfo(c *gin.Context) {

	userid := c.GetInt64("userid")
	//openid := c.GetString("openid")
	um := app.UserModel{}
	app.GetDb().Where("id = ?", userid).Get(&um)

	isBind := false
	if len(um.MihoyoAccountId) > 0 && len(um.MihoyoWebToken) > 0 {
		isBind = true
	}

	//判断今日是否签到
	isSign := false
	time.Now().Format("2006-01-02")
	t, _ := time.ParseInLocation("2006-01-02", time.Now().Format("2006-01-02"), time.Local)
	if um.SignTime >= t.Unix() {
		isSign = true
		//如果今日已签到的情况下,直接进入redis缓存!

	}

	//直接返回这个用户的个人信息
	c.JSON(200, gin.H{
		"status": 0,
		"msg":    "ok",
		"data": gin.H{
			//返回用户部分基础数据
			"id": um.Id,
			//绑定的游戏角色
			"isBind": isBind, //是否绑定过用户
			"bindPlayer": gin.H{
				"serverName": um.ServerName,                                           //所在服务器
				"playerName": um.PlayerName,                                           //玩家名称
				"playerUId":  um.PlayerUid,                                            //玩家UId
				"bindTime":   time.Unix(um.BindTime, 0).Format("2006-01-02 15:04:05"), //绑定时间
				"signTime":   time.Unix(um.SignTime, 0).Format("2006-01-02 15:04:05"), //上次签到时间
				"st":         um.SignTime,
				"isSign":     isSign,
				"totalSign":  um.TotalSign, //累计签到天数
			},
		},
	})
}

func GetToken(c *gin.Context) {
	token := c.Query("token")
	if helper.IsEmpty(token) {
		app.NewException("令牌错误!")
	}

	field := fmt.Sprintf("genshin_tmp_token:%s", token)

	_Token, err := app.GetRDB().Get(app.GetCtx(), field).Result()
	if err != nil {
		app.NewException("认证已过期或无效,请重新登录!")
	}
	app.GetRDB().Del(app.GetCtx(), field)
	c.JSON(200, gin.H{
		"status": 0,
		"msg":    "ok",
		"data": gin.H{
			"token": _Token,
		},
	})
}

//绑定角色
func BindPlayer(c *gin.Context) {

	accountId := c.PostForm("accountId")
	cookieToken := c.PostForm("cookieToken")

	//1.基础验证数据
	if helper.IsEmpty(accountId) || helper.IsEmpty(cookieToken) {
		app.NewException("AccountId或CookieToken不能为空")
	}
	userid := c.GetInt64("userid")

	umm := app.UserModel{}
	app.GetDb().Where("id = ?", userid).Get(&umm)

	//绑定间隔不能少于5秒
	if umm.BindTime+5 > time.Now().Unix() {
		app.NewException("绑定过于频繁,请稍后再试!")
	}

	genshin := api.NewGenshinApi()
	cookie := fmt.Sprintf("account_id=%s;cookie_token=%s", accountId, cookieToken)

	player, err := genshin.GetPlayerInfo(cookie)
	if err != nil {
		app.NewException(fmt.Sprintf("获取游戏角色失败:%s", err.Error()))
	}

	um := new(app.UserModel)
	um.MihoyoWebToken = cookieToken
	um.MihoyoAccountId = accountId
	um.ServerName = player.ServerName
	um.PlayerUid = player.GameUid
	um.PlayerName = player.NickName
	um.BindTime = time.Now().Unix()
	//修改数据库

	if _, err := app.GetDb().
		Cols("account_id", "web_token",
			"player_name", "server_name", "player_id", "bind_time",
		).
		Where("id = ?", userid).
		Update(um); err != nil {
		fmt.Println("绑定失败:", err.Error())
		app.NewException("绑定失败,系统错误!")
	}
	//绑定成功后,直接调用一次接口
	go func(player *api.GenshinPlayer) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("运行及时签到失败: %#v \n", err)
			}
		}()
		signStatus, err := genshin.RunSign(player.GameUid, cookie)
		isUpdate := false
		isNotifySign := false
		if signStatus == 0 {
			//fmt.Println(player.NickName,":今日签到成功")
			isUpdate = true
			isNotifySign = true
		} else if signStatus == 1 {
			//fmt.Println(player.NickName,":今日已签到,无需重复签到")
			isUpdate = true
		} else {
			fmt.Println(player.NickName, ":绑定签到运行失败", err)
		}

		//更新数据库为签到时间
		if isUpdate {
			info, err := genshin.GetPlayerSignInfo(player.GameUid, cookie)
			if err == nil {
				um := new(app.UserModel)
				um.SignTime = time.Now().Unix()
				um.TotalSign = info.TotalSignDay
				//修改数据库

				if _, err := app.GetDb().
					Cols("sign_time", "total_sign").
					Where("id = ?", userid).
					Update(um); err != nil {
					fmt.Println("更新数据失败:", err.Error())
				} else {
					if isNotifySign {
						bot := api.GetQQBot()
						notifyMsg := fmt.Sprintf("原神米游社%s签到成功列表", time.Now().Format("2006-01-02"))
						notifyMsg += fmt.Sprintf("\n[%d天]%s(%s)", info.TotalSignDay, player.NickName, player.GameUid)

						for _, g := range helper.GetConfig().QQBot.SignNotifyGroup {
							bot.SendMessage(g, []string{
								notifyMsg,
							})
						}
					}
				}
			}
		}
	}(player)

	t := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf("[%s] -> 来自【%s】的旅行者“%s”(%s) 绑定自动签到成功! \n", t, player.ServerName, player.NickName, player.GameUid)

	//发送绑定通知到群内
	bot := api.GetQQBot()
	for _, g := range helper.GetConfig().QQBot.BindNotifyGroup {
		bot.SendMessage(g, []string{
			msg,
		})
	}

	c.JSON(200, gin.H{
		"status": 0,
		"msg":    "绑定成功",
		"data": gin.H{
			"serverName": player.ServerName,
			"playerUid":  player.GameUid,
			"playerName": player.NickName,
		},
	})

}
