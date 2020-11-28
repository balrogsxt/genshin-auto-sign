package task

import (
	"fmt"
	"github.com/balrogsxt/genshin-auto-sign/api"
	"github.com/balrogsxt/genshin-auto-sign/app"
	"github.com/balrogsxt/genshin-auto-sign/helper"
	"time"
)

//每日签到任务
func RunSignTask(isFirst bool) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("执行任务计划发生错误!!!", err)
		}
	}()
	t, _ := time.ParseInLocation("2006-01-02", time.Now().Local().Format("2006-01-02"), time.Local)
	//如果时间处于每日凌晨,并且不为isFirst的话则不执行
	his := time.Now().Format("15:04")
	if his == "00:00" && !isFirst {
		return
	}
	if isFirst {
		fmt.Println("[主要任务计划]", time.Now().Format("2006-01-02 15:04:05"))
	}

	playerList := []app.UserModel{}

	err := app.GetDb().Where("account_id != '' and web_token != '' and sign_time < ?", t.Unix()).Find(&playerList)
	if err != nil {
		fmt.Println("查询失败: ", err.Error())
		return
	}
	signOkList := []app.UserModel{}

	genshin := api.NewGenshinApi()

	for _, item := range playerList {
		isOK := func(item *app.UserModel, isFirst bool) bool {
			max := 1
			if isFirst {
				max = 3 //每日首次处理允许3次容错率
			}

			defer func() {
				if err := recover(); err != nil {
					fmt.Println(item.Id, "签到意外错误: ", err)
				}
			}()

			cookie := fmt.Sprintf("account_id=%s;cookie_token=%s", item.MihoyoAccountId, item.MihoyoWebToken)
			for i := 0; i < max; i++ {
				player, err := genshin.GetPlayerInfo(cookie)
				if err != nil {
					fmt.Println(item.Id, "->", "获取米游社信息失败:", err.Error())
					continue
				}
				signInfo, err := genshin.GetPlayerSignInfo(player.GameUid, cookie)
				if err != nil {
					fmt.Println(item.Id, "->", "获取米游社签到信息失败:", err.Error())
					continue
				}
				if !signInfo.IsSign {
					//未签到
					status, err := genshin.RunSign(player.GameUid, cookie)
					if status == 0 {
						//签到成功
						signInfo.TotalSignDay += 1
						item.TotalSign = signInfo.TotalSignDay
						fmt.Printf("[%s]%s(%s) 签到成功->当前累计签到%d天\n", player.ServerName, player.NickName, player.GameUid, signInfo.TotalSignDay)
					} else if status == 1 {
						//今日已签到
						item.TotalSign = signInfo.TotalSignDay
						fmt.Printf("[%s]%s(%s) 已经签到->当前累计签到%d天\n", player.ServerName, player.NickName, player.GameUid, signInfo.TotalSignDay)
					} else {
						//签到失败
						fmt.Println(item.Id, "->", "签到失败:", err.Error())
						continue
					}
				} else {
					item.TotalSign = signInfo.TotalSignDay
					fmt.Printf("[%s]%s(%s) 已经签到->当前累计签到%d天\n", player.ServerName, player.NickName, player.GameUid, signInfo.TotalSignDay)
				}

				//设置签到时间、累计天数
				um := new(app.UserModel)
				um.SignTime = time.Now().Unix()
				um.TotalSign = signInfo.TotalSignDay
				if _, err := app.GetDb().Where("id = ?", item.Id).Cols("sign_time", "total_sign").Update(um); err != nil {
					fmt.Println("更新签到信息失败: ", err.Error())
					continue
				}
				return true
			}
			return false
		}(&item, isFirst)

		if isOK {
			signOkList = append(signOkList, item)
		}

	}

	if isFirst {
		fmt.Println("[主要任务计划执行结束]", time.Now().Format("2006-01-02 15:04:05"))
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
