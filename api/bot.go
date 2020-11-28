package api

import (
	"errors"
	"fmt"
	"github.com/balrogsxt/genshin-auto-sign/helper"
	"github.com/imroc/req"
)

type QQBot struct {
	url     string //服务器地址
	qq      string //机器人QQ号
	key     string //密钥
	session string
}

func NewQQBot() *QQBot {
	q := new(QQBot)

	conf := helper.GetConfig().QQBot
	q.url = conf.Url
	q.qq = conf.QQ
	q.key = conf.Key

	return q
}

func (this QQBot) SendMessage(target interface{}, textList []string) {
	session, err := this.getSession()
	if err != nil {
		fmt.Println("获取session失败:", err)
		return
	}

	txtList := make([]map[string]interface{}, 0)
	for _, item := range textList {
		txtList = append(txtList, map[string]interface{}{
			"type": "Plain",
			"text": item,
		})
	}

	u := fmt.Sprintf("%s/sendGroupMessage", this.url)
	j := req.BodyJSON(map[string]interface{}{
		"sessionKey":   session,
		"target":       target,
		"messageChain": txtList,
	})
	res, err := req.Post(u, j)
	if err != nil {
		fmt.Println("发送失败:", err)
		return
	}
	json := struct {
		Code      int
		Msg       string
		MessageId int
	}{}
	if err := res.ToJSON(&json); err != nil {
		fmt.Println("发送失败:", err)
		return
	}
	if json.Code != 0 {
		fmt.Println("获取失败:" + json.Msg)
		return
	}
	fmt.Printf("[群消息]群ID:%s -> 发送成功 -> %d \n", target, json.MessageId)
}

func (this QQBot) getSession() (string, error) {
	if len(this.session) != 0 {
		return this.session, nil
	}
	session, err := this.auth()
	if err != nil {
		return "", err
	}
	//开通session
	if err := this.verify(session); err != nil {
		return "", err
	}
	this.session = session
	return session, nil
}
func (this *QQBot) verify(session string) error {
	u := fmt.Sprintf("%s/verify", this.url)
	j := req.BodyJSON(map[string]interface{}{
		"sessionKey": session,
		"qq":         this.qq,
	})
	res, err := req.Post(u, j)
	if err != nil {
		return err
	}
	json := struct {
		Code int
		Msg  string
	}{}

	if err := res.ToJSON(&json); err != nil {
		return err
	}
	if json.Code != 0 {
		return errors.New("获取失败:" + json.Msg)
	}
	return nil
}

//认证机器人
func (this *QQBot) auth() (string, error) {
	u := fmt.Sprintf("%s/auth", this.url)
	j := req.BodyJSON(map[string]interface{}{
		"authKey": this.key,
	})
	res, err := req.Post(u, j)
	if err != nil {
		return "", err
	}
	json := struct {
		Code    int
		Msg     string
		Session string
	}{}

	if err := res.ToJSON(&json); err != nil {
		return "", err
	}
	if json.Code != 0 || len(json.Session) == 0 {
		return "", errors.New("获取失败:" + json.Msg)
	}
	return json.Session, nil
}
