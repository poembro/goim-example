
 # demo 
 
 ### 流程

- 用户请求POST /open/adduser  from-table: shop_id=1320000得到:
```
{
    "created_at": "2022-04-29 15:54:24",
    "device_id": "7241d0deb3fcf45dda85901acb59b1f1",
    "face": "http://img.touxiangwu.com/2020/3/uq6Bja.jpg",
    "nickname": "user193610",
    "platform": "web",
    "pushurl": "http://localhost:8090/open/push?&platform=web",
    "referer": "http://192.168.84.168:8083/im.html?shop_id=13200000000",
    "remote_addr": "192.168.84.168",
    "room_id": "d339a209ccbaca713fa5407a79a3c17d",
    "shop_face": "https://img.wxcha.com/m00/86/59/7c6242363084072b82b6957cacc335c7.jpg",
    "shop_id": "13200000000",
    "shop_name": "shop13200000000",
    "suburl": "ws://localhost:7923/ws",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNDA1NDg3Mjk5OTU5MTkzNjEwIiwiZGV2aWNlX2lkIjoiNzI0MWQwZGViM2ZjZjQ1ZGRhODU5MDFhY2I1OWIxZjEiLCJuaWNrbmFtZSI6InVzZXIxOTM2MTAiLCJleHAiOjE2ODI3NTQ4NjQsImlzcyI6ImdvbGFuZ3Byb2plY3QifQ.IBBpWzjMBTjhskA5G1BpXv5hOux4WIsXicMqOtlgmYI",
    "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.13; rv:96.0) Gecko/20100101 Firefox/96.0",
    "user_id": "405487299959193610"
}
```

- 建立websocket连接，走auth认证协议,将上面信息以json形式带给服务端

- 写入在线 
```
//    HSET userId_123 2000aa78df60000 {id:1,nickname:张三,face:p.png,}
//    SET  deviceId_2000aa78df60000  192.168.3.222
//    Zadd  shop_id  time() user_id

```


- 客服后台登录
```  
1.拉取所有在线
// zrevrange  shop_id  0, 50

2.建立连接，并订阅指定用户的房间

3.给指定用户发消息


```