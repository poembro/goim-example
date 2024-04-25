# examples 
一个运行在[goim](#)上的 客服聊天 软件。 当前demo 仅用来体现 goim 的实战应用

---

## 特点
- 简洁 
- 高性能 
- 采用 redis 做持久化

---

## 使用方式

### 一. 开三个控制台窗口 分别执行
-  make runjob 
-  make runlogic 
-  make runcomet

 

### 二. 新开浏览器标签页打开
- 浏览器打开 http://127.0.0.1:3111/_/



### 三. 推送消息
- curl "http://127.0.0.1:3111/goim/push/all?operation=1000" -d 'broadcast message'

