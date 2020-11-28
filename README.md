# 原神米游社自动签到小工具

## 简介
该项目用于处理米游社原神版块每日签到自动化处理功能,需要玩家提供米游社account_id与cookie_token实现

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
#redis密码
    password: 123456789
#redis选择库 
    index: 1
#mirai qq机器人配置
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
```


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
> 如果不需要自动部署到服务器,删除工作流下go.yml中`远程服务器部署`流程
```
//设置仓库secrets
SSH_HOST //ssh地址
SSH_KEY //ssh密钥
SSH_PORT //ssh端口
SSH_USER //ssh用户名
TOKEN //github token
```
