package api

import (
	"errors"
	"fmt"
	"github.com/balrogsxt/genshin-auto-sign/app"
	"github.com/balrogsxt/genshin-auto-sign/app/model"
	"github.com/balrogsxt/genshin-auto-sign/helper"
	"github.com/balrogsxt/genshin-auto-sign/helper/log"
	"github.com/imroc/req"
	"math/rand"
	"strings"
	"time"
)

const (
	AppVersion = "2.1.0"
	ClientType = "5"
	Referer    = "https://webstatic.mihoyo.com/bbs/event/signin-ys/index.html"
	UserAgent  = "Mozilla/5.0 (Linux; Android 5.1.1; f103 Build/LYZ28N; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/52.0.2743.100 Safari/537.36 miHoYoBBS/" + AppVersion
	ActId      = "e202009291139501"
)

type GenshinApi struct {
}

func NewGenshinApi() *GenshinApi {
	return new(GenshinApi)
}

//获取玩家角色信息
func (this *GenshinApi) GetPlayerInfo(cookie string) ([]*GenshinPlayer, int, error) {
	uri := "https://api-takumi.mihoyo.com/binding/api/getUserGameRolesByCookie?game_biz=hk4e_cn"
	header := req.Header{
		"Cookie": cookie,
	}
	res, err := req.Get(uri, header)
	if err != nil {
		return nil, 999, err
	}

	model := struct {
		RetCode int
		Message string
		Data    struct {
			List []*GenshinPlayer
		}
	}{}

	if err := res.ToJSON(&model); err != nil {
		return nil, 997, err
	}
	if model.RetCode != 0 {
		return nil, model.RetCode, errors.New(model.Message)
	}
	if len(model.Data.List) == 0 {
		return nil, 998, errors.New("未绑定玩家角色")
	}
	return model.Data.List, 0, nil
}

//获取玩家签到信息
func (this *GenshinApi) GetPlayerSignInfo(ServerRegion string, PlayerUid string, cookie string) (*GenshinSignInfo, int, error) {
	uri := fmt.Sprintf("https://api-takumi.mihoyo.com/event/bbs_sign_reward/info?act_id=%s&region=%s&uid=%s", ActId, ServerRegion, PlayerUid)
	header := req.Header{
		"Cookie": cookie,
	}
	res, err := req.Get(uri, header)
	if err != nil {
		return nil, 4001, err
	}

	model := struct {
		RetCode int
		Message string
		Data    GenshinSignInfo
	}{
		RetCode: -9999,
	}
	if err := res.ToJSON(&model); err != nil {
		return nil, 4002, err
	}
	if model.RetCode == 0 {
		return &model.Data, 0, nil
	} else {
		return nil, model.RetCode, errors.New(model.Message)
	}

}

//运行玩家签到 return 状态,是否远程调用,错误
func (this *GenshinApi) RunSign(player *model.PlayerSign, cookie string) (int, bool, error) {

	requestJson := map[string]interface{}{
		"act_id": ActId,
		"region": player.ServerRegion,
		"uid":    player.PlayerId,
	}
	uri := "https://api-takumi.mihoyo.com/event/bbs_sign_reward/sign"
	header := req.Header{
		"Content-Type":      "application/json",
		"x-rpc-device_id":   "F84E53D45BFE4424ABEA9D6F0205FF4A",
		"x-rpc-app_version": AppVersion,
		"x-rpc-client_type": ClientType,
		"Cookie":            cookie,
		"Referer":           Referer,
		"DS":                getDs(),
		"User-Agent":        UserAgent,
	}

	model := struct {
		RetCode int
		Message string
		Data    struct {
			Code string
		}
	}{}
	curlEnv, err := app.GetRDB().LPop(app.GetCtx(), "remoteApiPool").Result()
	if err == nil {
		app.GetRDB().RPush(app.GetCtx(), "remoteApiPool", curlEnv)
	}
	log.Info("")
	if helper.IsUrl(curlEnv) {
		log.Info("【当前使用远程Curl执行: %s】", curlEnv)
		h := make([]string, 0)
		for k, v := range header {
			h = append(h, fmt.Sprintf("%s:%s", k, v))
		}

		//远程环境运行
		param := req.Param{
			"url":    uri,
			"header": helper.JsonEncode(h),
			"data":   helper.JsonEncode(requestJson),
		}
		res, err := req.Post(curlEnv, param)
		if err != nil {
			return 500, true, err
		}
		resultStr, _ := res.ToString()
		log.Info("[远程返回详细] %s", resultStr)
		if res.Response().StatusCode != 200 {
			return 500, true, errors.New("请求失败: " + res.Response().Status)
		}
		if strings.Contains(resultStr, "Requests") {
			return 500, true, errors.New("请求服务器频繁")
		}
		if err := res.ToJSON(&model); err != nil {
			return 2, true, err
		}
		if model.RetCode == 0 && (model.Data.Code == "ok" || len(model.Data.Code) > 0) {
			return 0, true, nil
		} else if model.RetCode == -5003 {
			return 1, true, nil //已经签到过了
		} else {
			log.Info("签到异常: %s", resultStr)
			return 2, true, errors.New(model.Message + " 原始返回数据: " + resultStr)
		}
	} else {
		log.Info("【当前使用本地环境执行】")
		//本地环境运行
		json := req.BodyJSON(requestJson)
		res, err := req.Post(uri, header, json)
		if err != nil {
			return 2, false, err
		}
		resultStr, _ := res.ToString()
		log.Info("[本地返回详情] %s", resultStr)
		if strings.Contains(resultStr, "Requests") {
			return 500, false, errors.New("请求服务器频繁")
		}

		if err := res.ToJSON(&model); err != nil {
			return 2, false, err
		}
		if model.RetCode == 0 && (model.Data.Code == "ok" || len(model.Data.Code) > 0) {
			return 0, false, nil
		} else if model.RetCode == -5003 {
			return 1, false, nil //已经签到过了
		} else {
			log.Info("签到异常: %s", resultStr)
			return 2, false, errors.New(model.Message + " 原始返回数据: " + resultStr)
		}
	}

}

type GenshinPlayer struct {
	GameUid    string `json:"game_uid"`    //游戏ID
	NickName   string `json:"nickname"`    //游戏昵称
	ServerName string `json:"region_name"` //服务器名称
	Region     string `json:"region"`      //游戏服务器地址
}
type GenshinSignInfo struct {
	IsSign       bool `json:"is_sign"`        //今天是否签到
	TotalSignDay int  `json:"total_sign_day"` //本月累计签到天数
}

func getDs() string {
	t := time.Now().Unix()
	rand.Seed(time.Now().UnixNano())
	r := rand.Intn(999999)
	ms := fmt.Sprintf("salt=%s&t=%d&r=%d", helper.Md5(AppVersion), t, r)
	md5 := helper.Md5(ms)
	return fmt.Sprintf("%d,%d,%s", t, r, md5)
}
