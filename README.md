# PoleVPN Router

## PoleVPN Router 介绍
* polevpn router 是一种SD-WAN 技术，用来打通加速企业，不同办公区域，IDC 网络的一种网络技术，它工作在internet 协议之上
* 全程加密通信安全可靠
* 对业务透明，无感知，基于网络ip 协议层数据转发
* 安装容易，简单，golang 编写，没有依赖，高性能

## 系统架构图

![image](https://raw.githubusercontent.com/polevpn/polevpn_router/main/architecture.png)

## 系统架构说明
* polevpn router 跟 polevpn gateway 配合使用
* polevpn router 路由服务端用来转发gateway 传过来的ip层数据包
* polevpn gateway 是本地网络虚拟网关，其他机器或者网关可以配置路由到这台机器
* polevpn router 是采用 kcp，websocket 通信协议 来传输ip 数据报文

## polevpn router 安装使用
* 克隆项目git clone https://github.com/polevpn/polevpn_router.git
* cd polevpn_router
* go build
* nohup ./polevpn_router -configPath=./config.json &

## polevpn router 配置说明
```
{
    "kcp":{
        "listen":"0.0.0.0:443", //kcp 监听ip:port
        "enable":true //是否启用kcp 
    },
    "wss":{
        "listen":"0.0.0.0:443", //websocket 监听ip:port
        "enable":true, //是否启用websocket
        "cert_file":"./keys/server.crt", //websocket tls 通信证书
        "key_file":"./keys/server.key" //websocket tls 通信key
    },
    "shared_key":"!@#dFXemc$%*%^0K" //通信用的共享密钥
}
```

