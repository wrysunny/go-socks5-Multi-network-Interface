# go-socks5-Multi-network-Interface
go-socks5 Multi-network Interface 多网卡下的socks5服务

# 由来   
做渗透测试时，有线为内部办公网络（不能有攻击流量），无线为独立带宽，所有攻击的行为要走无线。   
因为默认有线的优先级比无线要高。所以默认情况下攻击流量走的是有线网口。   
不要跟我说 设置网络优先级、使用iptables ！ 机器应用太多，也懒得弄！    
就写了socks5 server 让其走自己设定的网口。   
目前用下来还没发现问题，有问题开issues。  

# 功能   
指定socks5服务发包走哪个网口  
默认不需账号密码认证（内部使用，也用不到鉴权）  
~~比默认socks5 server更安全。单项隔离？，只能从一个A网络代理到B网络，不能由B网络访问到A网络。（需要iptables配置）~~

# Usage  

Usage of main.exe:  
  -listen string  
        监听的socks5服务端口 (default "10880")  
  -outip string  
        想要从那个ip出口发包  
