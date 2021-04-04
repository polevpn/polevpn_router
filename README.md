# PoleVPN Router

## PoleVPN Router 介绍
* polevpn router 是一种SD-WAN 技术，用来打通加速企业，不同办公区域，IDC 网络的一种网络技术，它工作在internet 协议之上
* 全程加密通信安全可靠
* 对业务透明，无感知，基于网络ip 协议层数据转发
* 安装容易，简单，golang 编写，没有依赖，高性能
* 即可打通加速企业网络，也可作为VPN SERVER 使用

## 系统架构图

![image](https://raw.githubusercontent.com/polevpn/polevpn_router/main/architecture.png)

## 系统架构说明
* polevpn router 跟 polevpn gateway 配合使用
* polevpn router 作为整个虚拟路由系统的服务端用来转发polevpn gateway 传过来的ip层数据包
* polevpn gateway 是本地网络虚拟网关,通过创建虚拟网卡从本机网络协议栈获取IP数据包,以及把从polevpn router 接收到的ip 数据包写入到本机的网络协议栈，其他机器或者本地路由器网关可以配置路由到这台机器
* polevpn router跟polevpn gateway 之间采用 kcp，websocket 通信协议 来传输ip 数据报文

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
## 常见应用案例
- 案例1：某公司总部在广州，目前发展东南亚业务，在新加坡拥有一个分公司，东南亚业务的服务器放在新加坡区域的aws
    -  新加坡分公司需要跟广州总公司办公网络打通，而且需要比较好的网络延迟
    -  广州总公司的运维团队，需要远程维护新加坡区域的aws,同样需要低延迟网络
    -  由于GFW，国内出口网络带宽小的等因素，无论从广州，还是新加坡访问对方网络，ping 值都在260ms 以上，而且丢包严重，基本满足不了快速访问的需求
    -  通过采用低成本的cn2 网络出口（腾讯云或者阿里云的轻量海外服务器），每个月只需要几十块钱就可以，拥有30Mbps 的优质海外访问专线
    -  本案例可以在腾讯云或者阿里云轻量服务器上安装polevpn router，在广州总部网络,新加坡分公司网络，aws网络分别 安装polevpn gateway,就可以轻松打通这三个网络，同时获取50ms 左右的低延迟网络体验
- 案例2：某公司部分业务放在云端，部分业务放在公司自己的机房（比如办公室），想要打通公司自己机房跟云端网络
    - 可以在云端任意一台机器安装polevpn router 跟polevpn gateway 
    - 公司自己机房任意一台机器安装polevpn gateway 
    - 然后配置好相应的路由，就可以打通公司机房网络跟云端网络，不需要使用复杂的专门的vpn 设备
  
