package controller

import (
	"fmt"
	"github.com/balrogsxt/genshin-auto-sign/api"
	"github.com/balrogsxt/genshin-auto-sign/app"
	"github.com/balrogsxt/genshin-auto-sign/app/model"
	"github.com/balrogsxt/genshin-auto-sign/helper"
	"github.com/balrogsxt/genshin-auto-sign/helper/log"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var bindCd map[int64]int64 //绑定冷却
func init() {
	bindCd = make(map[int64]int64, 0)
}

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

	um := model.User{}
	var userId int
	//验证是否注册过账户
	has, err := app.GetDb().Where("openid = ?", openid).Get(&um)
	if has == false {
		conf := helper.GetConfig()
		if conf.NewUser == false {
			NoRegisterText := conf.NoRegisterText
			if len(conf.NoRegisterText) == 0 {
				NoRegisterText = "当前暂未开放新用户使用...(以前登录过的用户可正常使用)"
			}

			log.Info("[新用户尝试注册被挡] OpenId: %s -> ", openid, NoRegisterText)

			app.NewException(NoRegisterText)
		}

		//账户未注册,先注册
		userModel := new(model.User)
		userModel.OpenId = openid
		userModel.CreateTime = time.Now().Unix()
		_, err := app.GetDb().Insert(userModel)
		if err != nil {
			log.Info("注册失败:", err.Error())
			app.NewException("账户注册失败...请稍后再试!")
		}
		log.Info("[新用户注册成功] OpenId: %s -> ", openid)

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

	pl := []model.Player{}
	if err := app.GetDb().Where("uid = ?", userid).Find(&pl); err != nil {
		app.NewException("获取绑定角色列表失败...")
	}

	um := model.User{}
	app.GetDb().Where("id = ?", userid).Get(&um)

	playerList := make([]gin.H, 0)
	t, _ := time.ParseInLocation("2006-01-02", time.Now().Format("2006-01-02"), time.Local)
	for _, p := range pl {
		isSign := false
		if p.SignTime >= t.Unix() {
			isSign = true
		}

		playerList = append(playerList, gin.H{
			"id":         p.Pid,
			"serverName": p.ServerName,                                           //服务器名称
			"playerId":   p.PlayerId,                                             //玩家ID
			"playerName": p.PlayerName,                                           //玩家名称
			"signTime":   p.SignTime,                                             //签到时间
			"bindTime":   time.Unix(p.BindTime, 0).Format("2006-01-02 15:04:05"), //绑定时间
			"isSign":     isSign,                                                 //今日是否签到
			"totalSign":  p.TotalSign,                                            //累计签到天数
		})
	}

	//直接返回这个用户的个人信息
	c.JSON(200, gin.H{
		"status": 0,
		"msg":    "ok",
		"data": gin.H{
			//返回用户部分基础数据
			"id":         um.Id,
			"email":      um.Email,
			"bindPlayer": playerList,
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

//绑定邮箱
func BindEmail(c *gin.Context) {
	email := c.PostForm("email")
	if !helper.IsEmail(email) {
		app.NewException("邮箱格式不正确")
	}

	userid := c.GetInt64("userid")

	um := new(model.User)
	um.Email = email
	//邮箱允许重复
	if _, err := app.GetDb().
		Cols("email").
		Where("id = ?", userid).
		Update(um); err != nil {
		log.Info("绑定失败:", err.Error())
		app.NewException("绑定失败,系统错误!")
	}
	c.JSON(200, gin.H{
		"status": 0,
		"msg":    "绑定邮箱成功",
	})
}

//取消绑定
func UnBindPlayer(c *gin.Context) {
	userid := c.GetInt64("userid")
	_pid := c.PostForm("pid")
	pid, err := strconv.ParseInt(_pid, 10, 64)
	if err != nil {
		app.NewException("参数错误")
	}
	p := model.Player{}
	has, err := app.GetDb().Where("id = ? and uid = ?", pid, userid).Get(&p)
	if err != nil {
		log.Info("[解除绑定]查询角色失败: %s", err.Error())
		app.NewException("系统错误,请稍后再试!")
	}
	if !has {
		app.NewException("没有找到这个角色,无法解除绑定!")
	}

	if _, err := app.GetDb().
		Where("id = ? and uid = ?", pid, userid).
		Delete(&model.Player{}); err != nil {
		log.Info("取消绑定失败:", err.Error())
		app.NewException("取消绑定失败,系统错误!")
	}
	log.Info("[解除绑定] [%s]%s(%s)", p.ServerName, p.PlayerName, p.PlayerId)

	c.JSON(200, gin.H{
		"status": 0,
		"msg":    "取消绑定成功",
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

	userCd, has := bindCd[userid]
	if has {
		if userCd+5 > time.Now().Unix() {
			app.NewException("绑定过于频繁,请稍后再试!")
		}
	}

	genshin := api.NewGenshinApi()
	cookie := fmt.Sprintf("account_id=%s;cookie_token=%s", accountId, cookieToken)

	players, _, err := genshin.GetPlayerInfo(cookie)
	if err != nil {
		app.NewException(fmt.Sprintf("获取游戏角色失败:%s", err.Error()))
	}

	if len(players) == 0 {
		app.NewException("当前米游社账户可能暂未绑定任何游戏角色!!!")
	}
	bindPlayerList := make([]gin.H, 0)

	db := app.GetDb().NewSession()

	//防止滥用,绑定的时候移除当前账号下绑定的重新设置
	if _, err := db.Where("uid = ?", userid).Delete(&model.Player{}); err != nil {
		log.Info("删除已绑定的失败: %s", err.Error())
		app.NewException("绑定操作失败,系统异常")
	}
	//绑定到当前用户
	for _, player := range players {

		signInfo, status, err := genshin.GetPlayerSignInfo(player.Region, player.GameUid, cookie)
		if err != nil {
			log.Info("[绑定角色] [%s]%s(%s) -> 状态码: %d -> 获取签到信息失败: %s", player.ServerName, player.NickName, player.GameUid, status, err.Error())
			continue
		}

		newPlayer := new(model.Player)
		newPlayer.Uid = userid
		newPlayer.PlayerName = player.NickName
		newPlayer.ServerRegion = player.Region
		newPlayer.ServerName = player.ServerName
		newPlayer.BindTime = time.Now().Unix()
		newPlayer.PlayerId = player.GameUid
		newPlayer.TotalSign = signInfo.TotalSignDay
		if signInfo.IsSign {
			newPlayer.SignTime = time.Now().Unix()
		}

		if _, err := db.Insert(newPlayer); err == nil {
			bindPlayerList = append(bindPlayerList, gin.H{
				"player_id":   player.GameUid,
				"player_name": player.NickName,
				"server_name": player.ServerName,
			})
		}
	}
	if len(bindPlayerList) == 0 {
		app.NewException("绑定失败:没有绑定成功任何游戏角色...")
	}

	um := new(model.User)
	um.MihoyoWebToken = cookieToken
	um.MihoyoAccountId = accountId

	if _, err := db.
		Cols("account_id", "web_token").
		Where("id = ?", userid).
		Update(um); err != nil {

		db.Rollback()
		log.Info("绑定失败:", err.Error())
		app.NewException("绑定失败,系统错误!")

	} else {
		db.Commit()
	}
	bindCd[userid] = time.Now().Unix()

	//绑定成功后,直接调用一次接口
	title := "【游戏角色绑定成功】"
	notifyMsg := title
	for _, item := range bindPlayerList {
		notifyMsg += fmt.Sprintf("\n[%s]%s(%s)", item["server_name"], item["player_name"], item["player_id"])
	}
	if title != notifyMsg {
		log.Info(notifyMsg)
		//发送绑定通知到群内
		bot := api.GetQQBot()
		for _, g := range helper.GetConfig().QQBot.BindNotifyGroup {
			bot.SendMessage(g, []string{
				notifyMsg,
			})
		}
	}

	c.JSON(200, gin.H{
		"status": 0,
		"msg":    "绑定成功",
		"data":   bindPlayerList,
	})

}
