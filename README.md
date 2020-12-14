# 原神米游社自动签到小工具-服务端API

## 在线食用
[https://genshin.acgxt.com](https://genshin.acgxt.com)

## 简介
本项目主要是服务端API功能,网页方面需要自行实现或copy上面在线地址~,该项目用于处理米游社原神版块每日签到自动化处理功能,需要玩家提供米游社account_id与cookie_token实现,为了防止滥用,每一个QQ账号只允许绑定一个米游社账户

## 任务计划
项目运行后,每5分钟、每天凌晨0点都会触发一次签到操作,没有签到的用户将被处理【需要配置task配置】

## 配置
> 在根目录创建`config.yaml`文件,配置采用yaml配置文件 
```
#运行模式 release=生产环境 test=测试环境 debug=调试环境
run_mode: release
#服务绑定地址
http_host: 0.0.0.0
#服务运行端口
http_port: 8081
#授权成功同步回调地址
redirect_token_url: https://genshin.xt.com/?token=%token%
#qq授权信息
qqoauth:
    client_id: ***************
    client_secret: ***********
    redirect_uri: https://xxxx.com
#mysql配置
mysql:
    host: 127.0.0.1
    port: 端口号
    name: 数据库名称
    user: 用户名
    password: 密码
#redis配置
redis:
    host: 127.0.0.1
    port: 6379
    password: 123456789
    index: 1 #redis选择库
#mirai qq http插件机器人配置
qqbot:
    #接口地址
    url: http://*****:8080
    #机器人qq号
    qq: ******
    #机器人密钥
    key: ********
    #绑定成功后通知的群组(数组)
    bind_notify_group:
        - ****群号
    #签到成功后通知的群组(数组)
    sign_notify_group:
        - *****群号
    #cookie过期后通知的群组(数组)
    expire_notify_group:
        - ****群号
#由于米游社api请求过多会导致失败,这里可以配置远程curl接口调用
#数组api列表【文件位于项目目录curl目录下remote.php文件env >= php7.0】
curl_api:
  - https://*********/remote.php
  - https://*********/xxx/remote.php
#任务触发时间,第一个值将作为凌晨多次检测触发
task:
  - 10 0 0 * * *   #每日凌晨0点0分第十秒触发
  - 0 */5 * * * *  #每5分钟触发一次
```
> mysql数据库配置,就一张表 `user`
```
CREATE TABLE `user` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `openid` varchar(60) NOT NULL COMMENT 'openid',
  `account_id` varchar(50) DEFAULT NULL,
  `web_token` varchar(255) DEFAULT NULL,
  `create_time` int(10) DEFAULT NULL,
  `player_name` varchar(50) DEFAULT NULL,
  `server_name` varchar(50) DEFAULT NULL,
  `player_id` varchar(50) DEFAULT NULL,
  `bind_time` int(10) DEFAULT NULL COMMENT '绑定时间',
  `sign_time` int(10) DEFAULT '0' COMMENT '上一次签到事件',
  `total_sign` int(10) DEFAULT NULL COMMENT '累计签到天数',
  PRIMARY KEY (`id`),
  UNIQUE KEY `openid` (`openid`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;
```
## 服务端API(简略版)
> `GET` /login QQ账户登录

无参数
> `GET` /loginVerify QQ回调验证

|参数名|类型|说明|
|:----:|:----|----:|
|code|string|回调Code|
> `GET` /getToken 获取jwt密钥

|参数名|类型|说明|
|:----:|:----|----:|
|token|string|登录成功回调后的临时token|

> `GET` /logout  退出登录(需要提供authorization)

无参数
> `GET` /info 获取当前用户信息

无参数

> `POST` /bind 绑定米游社账户
> `POST` /unbind 解除绑定

|参数名|类型|说明|
|:----:|:----|----:|
|accountId|string|米游社accountid|
|cookieToken|string|米游社cookietoken|

## 本地编译&启动
```
//设置环境,linux、windows
SET GOOS=linux
//本地启动
go run main.go
//编译
go build main.go
```
## github actions自动编译
- 1.获取github token
- 2.设置项目仓库secrets
```
//设置仓库secrets
TOKEN //github token
```
