package task

import (
	"fmt"
	"github.com/balrogsxt/genshin-auto-sign/api"
	"github.com/balrogsxt/genshin-auto-sign/app"
	"github.com/balrogsxt/genshin-auto-sign/app/model"
	"github.com/balrogsxt/genshin-auto-sign/helper"
	"github.com/balrogsxt/genshin-auto-sign/helper/log"
	"strings"
	"time"
)

type PlayerSignInfo struct {
	ServerName string //服务器名称
	PlayerName string //玩家名称
	PlayerUid  string //玩家UId
	TotalSign  int    //签到天数
}

//每日签到任务
func RunSignTask(isFirst bool) {
	defer func() {
		if err := recover(); err != nil {
			log.Info("执行任务计划发生错误!!!", err)
		}
	}()
	t, _ := time.ParseInLocation("2006-01-02", time.Now().Local().Format("2006-01-02"), time.Local)
	//如果时间处于每日凌晨,并且不为isFirst的话则不执行
	his := time.Now().Format("15:04")
	if his == "00:00" && !isFirst {
		return
	}
	playerList := []model.PlayerSign{}

	if err := app.GetDb().
		Where("sign_time < ? ", t.Unix()).
		Join("INNER", "user", "user.id=player.uid").
		Cols("player.*", "user.account_id", "user.web_token", "user.email").
		Find(&playerList); err != nil {
		log.Info("查询需要签到的角色列表失败: %s", err.Error())
		return
	}
	if 0 >= len(playerList) {
		return
	}
	//签到成功的游戏玩家列表
	signList := []PlayerSignInfo{}

	runCount := 1 //可以执行的次数
	if isFirst {
		runCount = 3 //如果是每日凌晨0点则可以执行3次
	}

	genshin := api.NewGenshinApi()
	expireUser := make([]int64, 0)

	//每次运行执行
	for _, player := range playerList {
		log.Info("")
		log.Info("[%s]%s(%s)-----------------------------", player.ServerName, player.PlayerName, player.PlayerId)
		//检测是否过期的账号
		isExpire := false
		for _, uid := range expireUser {
			if uid == player.Uid {
				isExpire = true
				break
			}
		}
		if isExpire {
			continue
		}

		//组装必要cookie
		cookie := fmt.Sprintf("account_id=%s;cookie_token=%s", player.MihoyoAccountId, player.MihoyoWebToken)
		isErrRun := false
		for i := 0; i < runCount; i++ {
			//1.获取签到信息
			signInfo, retcode, err := genshin.GetPlayerSignInfo(&player, cookie)
			if err != nil {
				log.Info("[ %d ] -> [状态码: %d ] 获取米游社签到信息失败: %s", player.Uid, retcode, err.Error())

				if retcode == -100 {
					//登录失效,更新数据,通知用户
					expireUser = append(expireUser, player.Uid)
					go CookieExpireNotify(player)
					break
				}
				continue
			}
			//2.没有签到则进行签到操作
			if !signInfo.IsSign {
				status, isRemote, err := genshin.RunSign(&player, cookie)
				if status == 0 {
					//签到成功
					signInfo.TotalSignDay += 1
					log.Info("[ %d ] -> [%s]%s(%s) 今日签到成功 -> 当前累计签到%d天", player.Uid, player.ServerName, player.PlayerName, player.PlayerId, signInfo.TotalSignDay)
				} else if status == 1 {
					//今日已签到
					log.Info("[ %d ] -> [%s]%s(%s) 请勿重复签到 -> 当前累计签到%d天", player.Uid, player.ServerName, player.PlayerName, player.PlayerId, signInfo.TotalSignDay)
				} else {
					//签到失败
					log.Info("[ %d ] -> [%s]%s(%s) 签到发生错误 -> 状态码: %v -> 错误信息: %s", player.Uid, player.ServerName, player.PlayerName, player.PlayerId, signInfo.TotalSignDay, status, err.Error())

					if isRemote && isErrRun == false {
						//允许重新执行一次
						runCount += 1
						isErrRun = true
					}

					continue
				}
			} else {
				log.Info("[ %d ] -> [%s]%s(%s) 请勿重复签到 -> 当前累计签到%d天", player.Uid, player.ServerName, player.PlayerName, player.PlayerId, signInfo.TotalSignDay)
			}

			//3.保存签到信息
			signPlayerSave := new(model.Player)
			signPlayerSave.TotalSign = signInfo.TotalSignDay
			signPlayerSave.SignTime = time.Now().Unix()
			if _, err := app.GetDb().Where("id = ?", player.Pid).
				Cols("total_sign", "sign_time").
				Update(signPlayerSave); err != nil {
				log.Info("更新签到信息失败: %s", err.Error())
				continue
			}

			signList = append(signList, PlayerSignInfo{
				PlayerUid:  player.PlayerId,
				PlayerName: player.PlayerName,
				ServerName: player.ServerName,
				TotalSign:  signInfo.TotalSignDay,
			})
			break
		}
	}
	log.Info("")
	if len(signList) > 0 {
		log.Info("本次签到完成,累计: %d 人完成签到!", len(signList))
		notifyMsg := fmt.Sprintf("原神米游社%s签到成功列表", time.Now().Format("2006-01-02"))
		for _, item := range signList {
			notifyMsg += fmt.Sprintf("\n[%d天]%s(%s)", item.TotalSign, item.PlayerName, item.PlayerUid)
		}
		fmt.Println(notifyMsg)
		//发送签到通知到群内
		bot := api.GetQQBot()
		for _, g := range helper.GetConfig().QQBot.SignNotifyGroup {
			bot.SendMessage(g, []string{
				notifyMsg,
			})
		}
	}
}

//Cookie过期通知群组与邮箱
func CookieExpireNotify(player model.PlayerSign) bool {
	log.Info("【用户过期处理】 [%s]%s(%s)", player.ServerName, player.PlayerName, player.PlayerId)
	//1.删除cookie信息
	db := app.GetDb().NewSession()

	db.Begin()

	playerList := []model.PlayerSign{}
	if err := db.
		Where("uid = ?", player.Uid).
		Join("INNER", "user", "user.id=player.uid").
		Cols("player.*", "user.account_id", "user.web_token", "user.email").
		Find(&playerList); err != nil {
		log.Info("获取过期用户绑定的角色列表失败: %s", err.Error())
		db.Rollback()
		return false
	}
	if len(playerList) == 0 {
		//没有可以删除的角色
		db.Rollback()
		return true
	}

	u := new(model.User)
	if _, err := db.Where("id = ?", player.Uid).Cols("account_id", "web_token").Update(u); err != nil {
		db.Rollback()
		log.Info("删除过期用户的Cookie信息失败: %s", err.Error())
		return false
	}
	//2.删除账户下的绑定角色
	if _, err := db.Where("uid = ?", player.Uid).Delete(&model.Player{}); err != nil {
		db.Rollback()
		log.Info("删除过期用户下绑定的玩家角色失败: %s", err.Error())
		return false
	}
	db.Commit()
	db.Rollback()
	//3.发送通知相关
	notifyMsg := "【绑定已失效,请及时更换绑定】"
	for _, p := range playerList {
		notifyMsg += fmt.Sprintf("\n[%s]%s(%s)", p.ServerName, p.PlayerName, p.PlayerId)
	}
	log.Info(notifyMsg)
	go func() {
		//发送签到通知到群内
		bot := api.GetQQBot()
		for _, g := range helper.GetConfig().QQBot.SignNotifyGroup {
			bot.SendMessage(g, []string{
				notifyMsg,
			})
		}
		//发送邮件
		if helper.IsEmail(player.Email) {
			title := "阁下绑定的米游社账户已过期!"
			err := helper.SendEmail(player.Email, title, strings.ReplaceAll(notifyMsg, "\n", "<br/>"))
			if err == nil {
				log.Info("过期用户邮件通知成功 -> %s", player.Email)
			} else {
				log.Info("过期用户邮件通知失败: 地址: %s -> %s", player.Email, err.Error())
			}
		} else {
			log.Info("该用户未设置邮箱,无法绑定!")
		}
		//发送tg订阅推送

	}()
	return true
}
