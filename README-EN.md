# PoleVPN SD-WAN virtual routing system
* A network tunneling technology used to connect and accelerate different office areas of enterprises, IDC, cloud services, Internet of Things and other networks, making them work as if they were working in a local area network, and it works on top of the internet protocol
* The system is divided into polevpn router and [polevpn gateway](https://github.com/polevpn/polevpn_gateway)
* Full encryption communication is safe and reliable, based on kcp dtls, tcp tls protocol
* Transparent to business applications, no perception, data forwarding based on network IP protocol layer
* Easy to install, simple, written in golang, no dependencies, high performance
* You can get through the accelerated corporate network, and can also be used as a VPN SERVER
* Low-cost SD-WAN technology with low latency of private line
* Efficient anti-packet loss technology, FEC technology, lower network delay and smoother
* No need to install specific SD-WAN hardware devices

## System architecture diagram

![image](https://raw.githubusercontent.com/polevpn/polevpn_router/main/architecture.png)

## System architecture description
* The polevpn router works with the polevpn gateway
* The polevpn router is used as the server of the entire SD-WAN virtual routing system to forward the IP layer data packets transmitted by the polevpn gateway
* polevpn gateway is a virtual gateway of the local network. By creating a virtual network card, it obtains IP data packets from the local network protocol stack and sends them to the polevpn router, and writes the ip data packets received from the polevpn router to the local network protocol stack.
* Other machines or local routers can be configured to route to the machine where the polevpn gateway is located
* kcp, websocket communication protocol is used between polevpn router and polevpn gateway to transmit ip datagram

## PoleVPN Router Introduction
* polevpn router serves as the server of polevpn SD-WAN virtual routing system, providing routing services
* The communication protocol adopts the kcp protocol based on udp, and the kcp protocol is anti-packet loss, which has a great effect on delay-sensitive data
* Simultaneously supports tcp tls communication protocol, combined with bbr congestion algorithm, can achieve anti-GFW blocking at the protocol level, and low delay at the same time
* tcp communication encryption through tls, kcp communication encryption through dtls


## polevpn router installation and use
* Clone project git clone https://github.com/polevpn/polevpn_router.git
* cd polevpn_router
* go build
* nohup ./polevpn_router -configPath=./config.json &

## polevpn router configuration instructions
```
{
     "kcp": {
         "listen": "0.0.0.0:443", //kcp listens to ip:port
         "enable": true, //whether to enable kcp
         "cert_file":"./keys/server.crt", // dtls communication certificate
         "key_file":"./keys/server.key" //dtls communication key
     },
     "tls": {
         "listen": "0.0.0.0:443", //tls monitor ip:port
         "enable": true, //whether to enable tls
         "cert_file":"./keys/server.crt", // tls communication certificate
         "key_file":"./keys/server.key" // tls communication key
     },
     "key":"123456" //The key used to establish the connection
}
```
## Common Application Cases
- Case 1: A company is headquartered in Guangzhou, currently developing business in Southeast Asia, and has a branch in Singapore. The server for Southeast Asia business is placed in aws in Singapore
     - The Singapore branch needs to be connected to the office network of the Guangzhou head office, and requires a relatively good network delay
     - The operation and maintenance team of the Guangzhou head office needs to remotely maintain aws in the Singapore region, and also needs a low-latency network
     - Due to factors such as GFW and domestic export network bandwidth is small, whether you access the other party's network from Guangzhou or Singapore, the ping value is above 260ms, and the packet loss is serious, which basically cannot meet the needs of fast access
     - By using low-cost cn2 network export (Tencent Cloud or Aliyun's lightweight overseas server), it only costs tens of dollars per month, and has a high-quality overseas access line of 30Mbps
     - In this case, polevpn router can be installed on Tencent Cloud or Alibaba Cloud lightweight server, and polevpn gateway can be installed on the Guangzhou headquarters network, Singapore branch network, and aws network respectively, so that these three networks can be easily connected, and at the same time, the low-cost data of about 50ms can be obtained. Delayed network experience
- Case 2: A company puts part of its business in the cloud, and part of its business in the company's own computer room (such as an office), and wants to connect the company's own computer room with the cloud network
     - You can install polevpn router and polevpn gateway on any machine in the cloud
     - Install polevpn gateway on any machine in the company's own computer room
     - Then configure the corresponding routing, you can get through the company's computer room network and cloud network, without using complicated special vpn equipment
