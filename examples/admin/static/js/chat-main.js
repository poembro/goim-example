var app=new Vue({
    el: '#app',
    delimiters:["<{","}>"],
    data: {
        visible:false, //转接客服
        leftTabActive:"first", //左侧tab切换
        chatTitleType:"xxx",//聊天窗口标题类型
        chatTitle:"聊天窗口标题",
        rightTabActive:"userInfo", //右侧tab切换
        onlineUsers:[], //左侧tab在线用户列表 数据
        offlineUsers:[], //左侧tab在线用户列表 数据
        server:"ws://192.168.84.168:3102/sub",
        socket:null,
        messageContent:"", //请输入内容 
        msgList:[], // 消息列表  
        shop:{ // 客服配置信息
            shop_id : "13200000000",
            key:"web_13200000000",
            shop_name : "客服丽丽",
            shop_face : "https://img.wxcha.com/m00/86/59/7c6242363084072b82b6957cacc335c7.jpg", 
        }, 
        user:{ // 用户配置信息
            created_at: "-",
            key: "-",
            face: "http://img.touxiangwu.com/2020/3/uq6Bja.jpg",
            is_online: true,
            last_message: ["{\"mid\":405342717686579210, \"shop_id\":13200000000, \"type\":\"text\", \"msg\":\"3333\", \"room_id\":\"31bf19de027be5029716b460be4c16bd\", \"dateline\":1651132706, \"id\":\"1651132706180189000\"}"],
            nickname: "-",
            platform: "-",
            pushurl: "-",
            referer: "-",
            remote_addr: "-",
            room_id: "-",
            shop_face: "-",
            shop_id: "-",
            shop_name: "-",
            suburl: "-",
            unread: 1,
            user_agent: "-",
            mid:"-",
        },  
        userCount:0,
        userCurrentPage:1,
        userPageSize:10, 

        transKefuDialog:false, //转移客服
        otherKefus:[],   //客服
 
        ipBlacks:[{id:"1",remote_addr:"127.0.0.1"}], //黑名单
        sendDisabled:false,//发送
    },
    methods: {
        openUrl(url){
            window.open(url);
        },
        initConn() {//初始化websocket
            let token = this.shop
            this.socket = ReconnectingWebSocket.init(this.server, token, this);//创建Socket实例
        },
        OnMessage(msg) {
            let _this=this
            if (msg.msg == "") {
                return
            }
            //收到服务端消息
            let dst = {is_shop:false, face:"", name:"",msg:"", time:"", id:0}
            if (msg.mid == _this.shop.shop_id) 
            {
                dst.is_shop = true
                dst.face = _this.shop.shop_face
                dst.name = _this.shop.shop_name 
            }
            else 
            {
                dst.is_shop = false
                dst.face = _this.user.face
                dst.name = _this.user.nickname
            }
            dst.id = msg.id
            dst.msg = msg.msg 
            dst.dateline = msg.dateline
             
            _this.msgList.push(dst) //写入聊天记录
            _this.scrollBottom() //聊天记录网上拉
        },
        talkTo(user) {//接手客户
            this.user = user //用来判断,当前正在和谁聊天
            this.chatTitle = "正在与" + user.mid + "用户聊天...";
            
            // 清空上一个人的聊天记录
            this.msgList = []

            //获取历史消息列表
            this.getHistoryMsg(user.room_id);
            for(var i=0;i<this.onlineUsers.length;i++){
                if(this.onlineUsers[i].mid==user.mid){
                    this.$set(this.onlineUsers[i],'hidden_new_message',true);
                }
            }
            
            this.socket.sub(user.room_id) //订阅房间消息
            // 向房间内发一条欢迎语句
            let dst = {
                mid: this.shop.shop_id,
                shop_id: user.mid, 
                type: "text",
                msg: "+++++++正在接入,处理中...++", 
                room_id: user.room_id, 
                is_shop: true,
                dateline: 1645776553,
                id: "1645776553444263000"
            }
            this.socket.sendMsg(dst)
        },
        getHistoryMsg(room_id){ //获取历史消息列表
            let _this=this;
            console.log("获取历史消息列表")
            _this.sendAjax("/api/msg/list","post",{room_id:room_id}, function(dst){
                let result = dst.data
                let msg = {}
                for(let i= result.length - 1; i >=0; i--) {
                    msg = JSON.parse(result[i])
                    if (msg) _this.OnMessage(msg) 
                }
            })
        },
        closeUser(userId){//关闭访客
            console.log("关闭访客")
        },
        handleTabClick(tab, event){//处理tab切换
            let _this=this;
            if(tab.name=="first"){ 
                this.getOnlineUser() //在线用户tabs 获取在线游客信息
            }
            if(tab.name=="second"){ //离线用户
                this.getOfflineUser(1)
            }
            if(tab.name=="blackList"){ //黑名单
                this.listIpblack()
            }
            if(tab.name=="userInfo"){} //访客信息
        },
        listIpblack(){
            let _this=this 
            _this.ipBlacks = []
            let shop_id = _this.shop.shop_id  // 获取 黑名单列表 放到 _this.ipBlacks 
            _this.sendAjax("/api/ipblack/list","post",{shop_id:shop_id}, function(dst){
                var result = dst.data
               
                for(var i=0;i<result.length;i++){
                    _this.ipBlacks.push({id:i,remote_addr:result[i]})
                }
            })
        },
        addIpblack(ip){ // ip添加至黑名单
            let _this=this
            let shop_id = _this.shop.shop_id
            _this.sendAjax("/api/ipblack/add","post",{shop_id:shop_id, ip:ip}, function(dst){
                 console.log("ip添加至黑名单")
            })
        },
        delIpblack(ip){ // ip从黑名单删除
            let _this=this
            let shop_id = _this.shop.shop_id 
            _this.sendAjax("/api/ipblack/del","post",{shop_id:shop_id, ip:ip}, function(dst){
                console.log("ip从黑名单删除")
                _this.listIpblack()
            })
        },
        getOnlineUser(){//获取在线用户信息
            let _this=this
            let shop_id = _this.shop.shop_name
            console.log("参数 ",  _this.shop)

            _this.sendAjax("/api/shop/list","post",{shop_id:shop_id, typ:"online"}, function(dst){
                if (!dst) {
                    return console.log("暂无在线用户")
                }
                
                var result = dst.data
                //处理下 json字符串
                for(var i=0;i<result.length;i++){
                    let dst = result[i].last_message
                    if (dst.length > 0) {
                        let tmp = JSON.parse(dst[0])
                        result[i].last_message = tmp.msg
                    }
                }
                
                _this.onlineUsers=result; 
                for(var i=0;i<_this.onlineUsers.length;i++){
                    _this.$set(_this.onlineUsers[i],'hidden_new_message',true);
                }
            });
        },
        getOfflineUser(page){  //获取离线用户 
            let _this=this
            let shop_id = _this.shop.shop_name
            console.log("参数 ",page,  _this.shop)
            _this.sendAjax("/api/shop/list?page="+page, "post",{shop_id:shop_id, typ:"offline"}, function(dst){
                if (!dst) {
                    return console.log("暂无离线用户")
                }
                console.log(dst)
                var result = dst.data
                //处理下 json字符串
                for(var i=0;i<result.length;i++){
                    let dst = result[i].last_message
                    if (dst.length > 0) {
                        let tmp = JSON.parse(dst[0])
                        result[i].last_message = tmp.msg
                    }
                }

                _this.userCurrentPage = dst.page
                _this.offlineUsers = result
                _this.userCount = dst.total
                _this.userPageSize = dst.limit
            })

        },
             
        scrollBottom(){//滚到底部
            this.$nextTick(() => {
                $('.chatBox').scrollTop($(".chatBox")[0].scrollHeight);
            });
        },
        initJquery(){ //jquery
            this.$nextTick(() => {
                var _this=this;
                $(function () {  //展示表情/////////////
                });
            });
        },
        transKefu() { //转移客服
            this.transKefuDialog = true;
            var _this = this;
            this.sendAjax("/other_kefulist","get",{},function(result){
                _this.otherKefus=result;
            });
        },
        transKefuUser(kefu,userId) { //转移访客客服
            var _this = this;
            this.sendAjax("/trans_kefu","get",{kefu_id:kefu,mid:userId},function(result){
                //_this.otherKefus=result;
                _this.transKefuDialog = false
            });
        },
        sendMsg (){
            let _this = this;
            let dst = {
                mid: _this.shop.shop_id,
                shop_id: _this.user.mid,
                room_id:_this.user.room_id,
                type:'text',
                msg :_this.messageContent
            } 
            this.sendAjax("/api/msg/push", "post", dst, function(result){
                console.log(result)
            });
            _this.messageContent = "";
        },
        sendAjax(url,method,params,callback){
            let _this=this;
            $.ajax({
                type: method,
                url: url,
                data:JSON.stringify(params),
                headers: {
                    "token": _this.shop.token
                },
                error:function(res){
                    var data=JSON.parse(res);
                    if(!data.code || data.code!=200){
                        _this.$message({
                            message: data.msg,
                            type: 'error'
                        });
                    }
                },
                success: function(data) {
                    console.log("-->",url,"--",method,"---", data)
                    if (data.code!=200) {
                        _this.$message({
                            message: data.msg,
                            type: 'error'
                        });
                    } else {
                        callback(data);
                    }
                }
            });
        },
    },
    mounted() {},
    created: function () {
        let _this = this
        this.shop = getLocalStorage("shopinfo") 
        this.initJquery()
        this.initConn()

        this.getOnlineUser(); //获取在线游客信息
        setInterval(function(){
            _this.getOnlineUser();
        },12000);
    }
})
