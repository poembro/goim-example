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


### 二. 注册渠道/商户账号 
-  浏览器打开 注册账号 http://127.0.0.1:3111/admin/register.html  如: 账号 st3000 密码 123456 
-  浏览器打开 登录 http://127.0.0.1:3111/admin/login.html  如: 账号 st3000 密码 123456 


### 三. 新开浏览器标签页打开
- 浏览器打开 http://127.0.0.1:3111/front/?shop_id=st3000



### 四. 后台与第三步开的浏览器进行聊天
- 浏览器打开 http://127.0.0.1:3111/admin/main.html 