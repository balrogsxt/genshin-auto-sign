package api

import (
	"errors"
	"fmt"
	"github.com/balrogsxt/genshin-auto-sign/helper"
	"github.com/imroc/req"
	"math/rand"
	"time"
)

const (
	AppVersion = "2.1.0"
	ClientType = "5"
	Referer    = "https://webstatic.mihoyo.com/bbs/event/signin-ys/index.html"
	UserAgent  = "Mozilla/5.0 (Linux; Android 5.1.1; f103 Build/LYZ28N; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/52.0.2743.100 Safari/537.36 miHoYoBBS/" + AppVersion
	ActId      = "e202009291139501"
	Region     = "cn_gf01"
)

type GenshinApi struct {
}

func NewGenshinApi() *GenshinApi {
	return new(GenshinApi)
}

//获取玩家角色信息
func (this *GenshinApi) GetPlayerInfo(cookie string) (*GenshinPlayer, error) {
	uri := "https://api-takumi.mihoyo.com/binding/api/getUserGameRolesByCookie?game_biz=hk4e_cn"
	header := req.Header{
		"Cookie": cookie,
	}
	res, err := req.Get(uri, header)
	if err != nil {
		return nil, err
	}

	model := struct {
		RetCode int
		Message string
		Data    struct {
			List []GenshinPlayer
		}
	}{}

	if err := res.ToJSON(&model); err != nil {
		return nil, err
	}
	if model.RetCode != 0 {
		return nil, errors.New(model.Message)
	}
	if len(model.Data.List) == 0 {
		return nil, errors.New("未绑定游戏角色")
	}
	return &model.Data.List[0], nil
}

//获取玩家签到信息
func (this *GenshinApi) GetPlayerSignInfo(playerUid string, cookie string) (*GenshinSignInfo, error) {
	uri := fmt.Sprintf("https://api-takumi.mihoyo.com/event/bbs_sign_reward/info?act_id=e202009291139501&region=cn_gf01&uid=%s", playerUid)
	header := req.Header{
		"Cookie": cookie,
	}
	res, err := req.Get(uri, header)
	if err != nil {
		return nil, err
	}

	model := struct {
		RetCode int
		Message string
		Data    GenshinSignInfo
	}{}
	if err := res.ToJSON(&model); err != nil {
		return nil, err
	}
	return &model.Data, nil
}

//运行玩家签到
func (this *GenshinApi) RunSign(playerUid string, cookie string) (int, error) {
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
	json := req.BodyJSON(map[string]interface{}{
		"act_id": ActId,
		"region": Region,
		"uid":    playerUid,
	})
	res, err := req.Post(uri, header, json)
	if err != nil {
		return 2, err
	}

	model := struct {
		RetCode int
		Message string
		Data    struct {
			Code string
		}
	}{}

	if err := res.ToJSON(&model); err != nil {
		return 2, err
	}
	if model.RetCode == 0 && (model.Data.Code == "ok" || len(model.Data.Code) > 0) {
		return 0, nil
	} else if model.RetCode == -5003 {
		return 1, nil //已经签到过了
	} else {
		return 2, errors.New(model.Message)
	}
}

type GenshinPlayer struct {
	GameUid    string `json:"game_uid"`    //游戏ID
	NickName   string `json:"nickname"`    //游戏昵称
	ServerName string `json:"region_name"` //服务器名称
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
