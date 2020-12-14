package task

import (
	"fmt"
	"github.com/balrogsxt/genshin-auto-sign/api"
	"github.com/balrogsxt/genshin-auto-sign/app"
	"github.com/balrogsxt/genshin-auto-sign/helper"
	"github.com/balrogsxt/genshin-auto-sign/helper/log"
	"time"
)

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
	if isFirst {
		log.Info("[主要任务计划]", time.Now().Format("2006-01-02 15:04:05"))
	}

	playerList := []app.UserModel{}

	err := app.GetDb().Where("account_id != '' and web_token != '' and sign_time < ?", t.Unix()).Find(&playerList)
	if err != nil {
		log.Info("查询失败: %s", err.Error())
		return
	}
	signOkList := []app.UserModel{}

	genshin := api.NewGenshinApi()
	// todo 已知问题:用户过多后会导致api调用失败
	//每次运行执行
	for _, item := range playerList {
		status := func(item *app.UserModel, isFirst bool) int {
			max := 1
			if isFirst {
				max = 3 //每日首次处理允许3次容错率
			}

			defer func() {
				if err := recover(); err != nil {
					log.Info("用户ID: %#v -> 签到意外错误: %s", item.Id, err)
				}
			}()

			cookie := fmt.Sprintf("account_id=%s;cookie_token=%s", item.MihoyoAccountId, item.MihoyoWebToken)
			isRs := false
			for i := 0; i < max; i++ {
				if i >= 1 {
					time.Sleep(time.Second * 2) //如果出现错误一次则会延迟2秒执行
				}
				player, state, err := genshin.GetPlayerInfo(cookie)
				if state != 0 {
					log.Info("用户ID: %#v -> 获取米游社签到信息失败: %s", item.Id, err.Error())
					if state == -100 {
						//登录失效,清空cookie
						go func(item *app.UserModel) {
							//发送邮箱给玩家失效
							um := new(app.UserModel)
							um.MihoyoAccountId = ""
							um.MihoyoWebToken = ""
							um.ServerName = ""
							um.PlayerName = ""
							um.PlayerUid = ""
							um.BindTime = 0
							um.SignTime = 0
							um.TotalSign = 0
							if _, err := app.GetDb().Where("id = ?", item.Id).Cols("sign_time", "bind_time", "player_id", "player_name", "server_name", "total_sign", "account_id", "web_token").Update(um); err != nil {
								log.Info("更新过期数据失败: ", err.Error())
							}
							CookieExpireNotify(item)
						}(item)
					}
					continue
				}
				signInfo, err := genshin.GetPlayerSignInfo(player.GameUid, cookie)
				if err != nil {
					log.Info("%d -> 获取米游社签到信息失败: %s", item.Id, err.Error())
					continue
				}
				if !signInfo.IsSign {
					//if true {
					//未签到
					status, isRemote, err := genshin.RunSign(player.GameUid, cookie)
					if status == 0 {
						//签到成功
						signInfo.TotalSignDay += 1
						item.TotalSign = signInfo.TotalSignDay
						log.Info("[%s]%s(%s) 用户ID: %#v -> 签到成功 -> 当前累计签到%d天", player.ServerName, player.NickName, player.GameUid, item.Id, signInfo.TotalSignDay)
					} else if status == 1 {
						//今日已签到
						item.TotalSign = signInfo.TotalSignDay
						log.Info("[%s]%s(%s) 用户ID:%#v -> 已经签到 -> 当前累计签到%d天", player.ServerName, player.NickName, player.GameUid, item.Id, signInfo.TotalSignDay)
					} else {
						//签到失败
						log.Info("[签到失败] 用户ID:%#v -> 状态: %#v -> 错误信息: %#v ", item.Id, status, err)

						if isRemote && isRs == false {
							//远程执行的话,再运行一次接口
							max += 1
							isRs = true
							continue
						}
						//if status == 500 {
						//	//超时频繁,直接结束此次任务,放到下一次任务计划执行
						//	return 2
						//}

						continue
					}
				} else {
					item.TotalSign = signInfo.TotalSignDay
					log.Info("[%s]%s(%s) 已经签到->当前累计签到%d天", player.ServerName, player.NickName, player.GameUid, signInfo.TotalSignDay)
				}

				//设置签到时间、累计天数
				um := new(app.UserModel)
				um.SignTime = time.Now().Unix()
				um.TotalSign = signInfo.TotalSignDay
				if _, err := app.GetDb().Where("id = ?", item.Id).Cols("sign_time", "total_sign").Update(um); err != nil {
					log.Info("更新签到信息失败: ", err.Error())
					continue
				}
				return 0 //签到成功了
			}
			return 1 //执行失败了
		}(&item, isFirst)

		if status == 0 {
			signOkList = append(signOkList, item)
		}
		//else if status == 2 {
		//	//终结任务
		//	log.Info("[任务停止]检测到服务器请求过于频繁,本次任务终止,下一次任务计划继续执行!")
		//	break
		//}
		time.Sleep(time.Millisecond * 100)

	}

	if isFirst {
		log.Info("[主要任务计划执行结束]", time.Now().Format("2006-01-02 15:04:05"))
	}

	if len(signOkList) > 0 {
		notifyMsg := fmt.Sprintf("原神米游社%s签到成功列表", time.Now().Format("2006-01-02"))
		for _, item := range signOkList {
			notifyMsg += fmt.Sprintf("\n[%d天]%s(%s)", item.TotalSign, item.PlayerName, item.PlayerUid)
		}
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
func CookieExpireNotify(player *app.UserModel) {
	//发送签到通知到群内
	bot := api.GetQQBot()
	for _, g := range helper.GetConfig().QQBot.SignNotifyGroup {
		bot.SendMessage(g, []string{
			fmt.Sprintf("【%s】%s(%s)的绑定已过期,请及时重新绑定!", player.ServerName, player.PlayerUid, player.PlayerName),
		})
	}

	//helper.SendEmail(player.Email,"阁下的自动签到部署Cookie已过期,请重新部署设置!")
}
