package api

import (
	"errors"
	"fmt"
	"github.com/balrogsxt/genshin-auto-sign/helper"
	"github.com/imroc/req"
	"net/url"
	"time"
)

type QQAccessToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresTime  int64  `json:"expires_in`
	RefreshToken string `json:"refresh_token"`
	ErrorCode    int    `json:"error"`
	ErrorMsg     string `json:"error_description"`
}
type QQOpenId struct {
	ClientId  string `json:"client_id"`
	OpenId    string `json:"openid"`
	ErrorCode int    `json:"error"`
	ErrorMsg  string `json:"error_description"`
}

//QQ互联
type QQConnect struct {
	config *helper.Config
}

func NewQQ() *QQConnect {
	instance := new(QQConnect)
	instance.config = helper.GetConfig()
	return instance
}

//构建url
func (this *QQConnect) BuildLoginUrl() string {
	appid := this.config.QQOauth.ClientId
	redirect_uri := url.QueryEscape(this.config.QQOauth.RedirectUri)
	state := helper.Md5(fmt.Sprintf("%d", time.Now().UnixNano()/1e6))
	scope := "get_user_info,get_other_info,get_info"
	return fmt.Sprintf(`https://graph.qq.com/oauth2.0/authorize?response_type=code&client_id=%s&redirect_uri=%s&state=%s&scope=%s`, appid, redirect_uri, state, scope)
}

//验证code
func (this *QQConnect) GetAccessToken(code string) (*QQAccessToken, error) {
	param := req.QueryParam{
		"grant_type":    "authorization_code",
		"client_id":     this.config.QQOauth.ClientId,
		"client_secret": this.config.QQOauth.ClientSecret,
		"code":          code,
		"redirect_uri":  this.config.QQOauth.RedirectUri,
		"fmt":           "json",
	}
	res, err := req.Get("https://graph.qq.com/oauth2.0/token", param)
	if err != nil {
		return nil, err
	}
	json := new(QQAccessToken)
	if err := res.ToJSON(&json); err != nil {
		return nil, err
	}
	if json.ErrorCode != 0 {
		return nil, errors.New(json.ErrorMsg)
	}
	return json, nil
}

//获取QQ opneid
func (this *QQConnect) GetOpenId(accessToken string) (*QQOpenId, error) {
	u := "https://graph.qq.com/oauth2.0/me"
	param := req.QueryParam{
		"access_token": accessToken,
		"fmt":          "json",
	}
	res, err := req.Get(u, param)
	if err != nil {
		return nil, err
	}
	json := new(QQOpenId)

	if err := res.ToJSON(&json); err != nil {
		return nil, err
	}

	if json.ErrorCode != 0 {
		return nil, errors.New(json.ErrorMsg)
	}
	return json, nil
}
