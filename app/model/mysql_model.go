package model

type User struct {
	Id              int    `xorm:"pk autoincr"` //自增ID
	OpenId          string `xorm:"openid"`      //qqopenid
	MihoyoAccountId string `xorm:"account_id"`  //米游社账户Id
	MihoyoWebToken  string `xorm:"web_token"`   //米游社账户Token
	CreateTime      int64  `xorm:"create_time"` //注册时间
	Email           string `xorm:"email"`       //邮箱
}

type Player struct {
	Pid          int    `xorm:"pk autoincr 'id'"` //自增ID
	Uid          int64  //用户ID
	ServerRegion string `xorm:"server_region"` //所属服务器
	ServerName   string `xorm:"server_name"`   //服务器名称
	PlayerName   string `xorm:"player_name"`   //玩家名称
	PlayerId     string `xorm:"player_id"`     //玩家ID
	BindTime     int64  `xorm:"bind_time"`     //绑定时间
	SignTime     int64  `xorm:"sign_time"`     //签到时间
	TotalSign    int    `xorm:"total_sign"`    //累计签到时间
}

type PlayerSign struct {
	Player `xorm:"extends"`
	User   `xorm:"extends"`
}

func (PlayerSign) TableName() string {
	return "player"
}
