(function(win){
    var getQuery = function (name){
        var reg = new RegExp("(^|&)"+ name +"=([^&]*)(&|$)");
        var r = win.location.search.substr(1).match(reg);
        if(r!=null)return  unescape(r[2]); return null;
    }
    
    //存储localStorge
    var setLocalStorage = function (key, obj) {
        if (!navigator.cookieEnabled || typeof window.localStorage == 'undefined') {
            return false;
        }
        localStorage.setItem(key, JSON.stringify(obj));
        return true;
    }
    //读取localStorge
    var getLocalStorage = function (key) {
        if (!navigator.cookieEnabled || typeof window.localStorage == 'undefined') {
            return false;
        }
        var str = localStorage.getItem(key);
        if (!str) {
            return false;
        }
        return JSON.parse(str);
    }

    var _ = {} 
    _.init = function() {
        _.config.init() //必须最先执行
        _.face.init()  //工具处理 
        _.comet.init()
        _.upload.init()
    }

    _.sendAjax = function (url,method,params,callback,is_async) {
        let _this=this;
        $.ajax({
            type: method,
            url: url,
            dataType: 'json',
            contentType: 'application/json', 
            data:params,
            async: is_async,
            headers: {
                "token": _.config.options.token
            },
            error:function(res){
                var data=JSON.parse(res);
                if(data.code != 200){
                    _.alert(data.msg)
                }
            },
            success: function(data) {
                console.log("-->",url,"--",method,"---", data)
                if (data.code == 200 && callback) {
                    // _.alert(data.msg)
                    callback(data.data);
                } else if (data.result!=null){
                    callback(data.result);
                } else {
                    callback(data);
                }
            }
        });
    }

    _.config = {
        options : {
            room_id:'live://1000', //将消息发送到指定房间
           
            device_id:'1xxxx',  
            mid: '663291537152950273',
            nickname:'随机用户001',
            face:'/static/wap/img/portrait.jpg',

            shop_name:'杂货铺老板', 
            shop_id:0,
            shop_face:'/static/wap/img/portrait.jpg',
            platform:'web',
            suburl : "ws://192.168.3.222:9999/sub", 
            pushurl:"http://192.168.3.222:9999/open/push",
        },
        init:function (){
            var self = this  
            var __KEY__ = "_me_" 
            var old = getLocalStorage(__KEY__) 
            if (old) {
                self.options = old
                self.handleTitle(self.options.shop_name) 
                _.websocket.init() 

                return 
            }
            
            var shop_id = getQuery("shop_id")
            _.sendAjax("/api/user/create", "GET", {shop_id:shop_id}, function(dst){
                self.options = dst
                setLocalStorage(__KEY__, dst)
                self.handleTitle(dst.shop_name) 
                _.websocket.init() 

                return 
            }, false);
          
        }, 
        handleTitle : function(title) {
            $("#top_title").html(title) 
        }
    }

    _.comet = {
        init:function (){
            var self = this
            $('#console_box_input').bind('focus',function(){
               _.scrollTop()   
            }).bind('blur', function(){
               _.scrollTop()
            })

            $('#console_box_right').click(function(){
                self.send(null)
                $('#face_box').addClass("hide")
            })

            $(document).keyup(function(event){
                if (event.keyCode=="13" && event.shiftKey != 1){ //13表示回车键的代码
                    self.send(null)
                }
            })
            return true
        }, 
        send : function(msgtype) { //发送消息
            var console_box_input = $('#console_box_input')
            var msg = console_box_input.val() //拿到输入框内容
            var marr = msg.split("\n")
            var arr= marr.filter(function(x){ return x != "" }); 
            msg = arr.join("\r\n")
            var data = {
                type : 'text',
                msg : msg,
                room_id : _.config.options.room_id,
                mid: _.config.options.mid,
                nickname: _.config.options.nickname,
                face: _.config.options.face,
                shop_id: _.config.options.shop_id,
                shop_name: _.config.options.shop_name, 
            }

            if (msgtype && msgtype.length > 5) {
                data["type"]  = 'image'  
                data["msg"]  = msgtype 
            }

            if (!data["msg"] || $.trim(data["msg"]) == "" || (data["msg"].length < 1 ) || (data["msg"].length > 65536 ) ) { 
                _.alert('请输入内容再发送')
                return false 
            }

            var room_id = _.config.options.room_id
            var mid = _.config.options.mid 
            var nickname = _.config.options.nickname
            var face = _.config.options.face
            var shop_id = _.config.options.shop_id
            var shop_name = _.config.options.shop_name
            var url = _.config.options.pushurl + "?room_id=" + room_id+ "&mid=" + mid + "&nickname="+nickname+"&face="+face+"&shop_id="+shop_id+"&shop_name="+shop_name
            url = encodeURI(url)

            _.sendAjax(url, "POST", JSON.stringify(data), function(result){
                return 
            }, true);
      
            console_box_input.val("") 
            return true
        }
    }
 
    const rawHeaderLen = 16
    const packetOffset = 0
    const headerOffset = 4
    const verOffset = 6
    const opOffset = 8
    const seqOffset = 12 

    const MAX_CONNECT_TIMES = 10 //最大重连次数
    const DELAY = 7500          //每隔15秒连一次 
    _.websocket = {
        msgSeq : 0,
        ws : null,
        textDecoder : null,
        textEncoder : null,
        heartbeatInterval : null,  //定时器句柄 
        init : function(){
            var self = this
            self.textDecoder = new TextDecoder()
            self.textEncoder = new TextEncoder()

            self.createConnect(MAX_CONNECT_TIMES, DELAY)
        }, 
        createConnect : function (max, delay) {
            var self = this
            if (max === 0) {
                return
            }

            var ws = new WebSocket(_.config.options.suburl)
            ws.binaryType = 'arraybuffer'
            ws.onopen = function() {
                self.auth(ws) 
                var ishide = $("#doconfig")
                if (ishide) ishide.click();
            }

            ws.onmessage = function(evt) {
                var data = evt.data
                var dataView = new DataView(data, 0)
                var packetLen = dataView.getInt32(packetOffset)
                var headerLen = dataView.getInt16(headerOffset)
                var ver = dataView.getInt16(verOffset)
                var op = dataView.getInt32(opOffset)
                var seq = dataView.getInt32(seqOffset) 
                console.log("receiveHeader: packetLen=" + packetLen, "headerLen=" + headerLen, "ver=" + ver, "op=" + op, "seq=" + seq)
                switch(op) {
                    case 8: // 认证成功的结果
                        self.sub(ws, _.config.options.room_id)
                        // auth reply ok 
                        self.heartbeat(ws) 
                        self.heartbeatInterval = setInterval(function(){
                            self.heartbeat(ws)
                        }, 27 * 1000)
                        break
                    case 3: //心跳包成功的结果
                        //console.log("receive: heartbeat") 
                        break

                    case 15: //订阅房间的结果
                        //console.log("receive: sub")  
                        if (max === MAX_CONNECT_TIMES) {
                            self.sync(ws)
                        }
                        break  
                    case 17: // 取消订阅的结果 
                        break
                    case 19: //sync 同步历史消息
                        //var msgBody = self.textDecoder.decode(data.slice(headerLen, packetLen))
                        //console.log("receive 19 : 同步历史消息 ver=" + ver + " op=" + op + " seq=" + seq + " message=" + msgBody)
                        //self.syncMsgReceived(ws, msgBody)
                        break
                    case 21: //消息偏移上报的结果
                        //console.log("receive: ack") 
                        break
                    case 9: 
                        // batch message 原始消息 比如 TLV  中的 V (body体)又是一个 TLV 结构
                        var offset = rawHeaderLen 
                        for (; offset<data.byteLength; offset+=packetLen) {
                            // parse
                            var packetLen = dataView.getInt32(offset)
                            var headerLen = dataView.getInt16(offset+headerOffset)
                            var ver = dataView.getInt16(offset+verOffset)
                            var op = dataView.getInt32(offset+opOffset)
                            var seq = dataView.getInt32(offset+seqOffset)
                            var msgBody = self.textDecoder.decode(data.slice(offset+headerLen, offset+packetLen))
                            // callback 
                            //console.log("receive1: ver=" + ver + " op=" + op + " seq=" + seq + " message=" + msgBody)
                            self.messageReceived(ws, msgBody) 
                        }
                        break
                    case 5:
                        var msgBody = self.textDecoder.decode(data.slice(headerLen, packetLen))
                        console.log("receive 5: ver=" + ver + " op=" + op + " seq=" + seq + " message=" + msgBody)
                        self.messageReceived(ws, msgBody)
                        break
                    default:
                        // TODO
                        console.log("未知消息响应: packetLen=" + packetLen, "headerLen=" + headerLen, "ver=" + ver, "op=" + op, "seq=" + seq)
                }
            }

            ws.onclose = function() {
                if (self.heartbeatInterval) clearInterval(self.heartbeatInterval);
                setTimeout(reConnect, delay)
                _.alert("连接异常...")
            }
            function reConnect() {
                self.createConnect(--max, delay * 2)
            } 
        }, 
        sub : function (ws, room_id) { //订阅房间
            var self = this
            var headerBuf = new ArrayBuffer(rawHeaderLen) //分配16个固定元素大小
            var headerView = new DataView(headerBuf, 0) //读写时手动设定字节序的类型
            var bodyBuf = self.textEncoder.encode(room_id)
            headerView.setInt32(packetOffset, rawHeaderLen + bodyBuf.byteLength) //写入从内存的第0个字节序开始  值为16
            headerView.setInt16(headerOffset, rawHeaderLen) //写入从内存的第4个字节序开始  值为16
            headerView.setInt16(verOffset, 1)  //写入从内存的6第个字节序开始，值为1     省去的第三个参数: true为小端字节序，false为大端字节序 不填为大端字节序
            headerView.setInt32(opOffset, 14)   //写入从内存的8第个字节序开始，值为2
            headerView.setInt32(seqOffset, 1)  //写入从内存的12第个字节序开始，值为1 
            var flag = ws.send(self.mergeArrayBuffer(headerBuf, bodyBuf))
            return flag
        },
        sync : function (ws) {
            var self = this
            var dst = {
                page:1,
                op : _.config.options.accepts[0],
                key : _.config.options.key,
                room_id : _.config.options.room_id  
            }
            var token = JSON.stringify(dst)

            var headerBuf = new ArrayBuffer(rawHeaderLen)  
            var headerView = new DataView(headerBuf, 0)  
            var bodyBuf = self.textEncoder.encode(token)

            headerView.setInt32(packetOffset, rawHeaderLen + bodyBuf.byteLength)  
            headerView.setInt16(headerOffset, rawHeaderLen)  
            headerView.setInt16(verOffset, 1)  
            headerView.setInt32(opOffset, 18)
            headerView.setInt32(seqOffset, 1)  
         
            return ws.send(self.mergeArrayBuffer(headerBuf, bodyBuf))
        },
        messageAck : function (ws, key, roomId, id) {
            var self = this
            var dst = {key :key, room_id :roomId, id:id}
            var token = JSON.stringify(dst)
            var headerBuf = new ArrayBuffer(rawHeaderLen) //分配16个固定元素大小
            var headerView = new DataView(headerBuf, 0) //读写时手动设定字节序的类型
            var bodyBuf = self.textEncoder.encode(token)
            headerView.setInt32(packetOffset, rawHeaderLen + bodyBuf.byteLength) //写入从内存的第0个字节序开始  值为16
            headerView.setInt16(headerOffset, rawHeaderLen) //写入从内存的第4个字节序开始  值为16
            headerView.setInt16(verOffset, 1)  //写入从内存的6第个字节序开始，值为1     省去的第三个参数: true为小端字节序，false为大端字节序 不填为大端字节序
            headerView.setInt32(opOffset, 20)   //写入从内存的8第个字节序开始，值为2
            headerView.setInt32(seqOffset, 1)  //写入从内存的12第个字节序开始，值为1 
            var flag = ws.send(self.mergeArrayBuffer(headerBuf, bodyBuf))
            return flag
        },
        heartbeat : function (ws) {
            var headerBuf = new ArrayBuffer(rawHeaderLen)  
            var headerView = new DataView(headerBuf, 0)  
            headerView.setInt32(packetOffset, rawHeaderLen)  
            headerView.setInt16(headerOffset, rawHeaderLen)  
            headerView.setInt16(verOffset, 1)  
            headerView.setInt32(opOffset, 2)
            headerView.setInt32(seqOffset, 1)  
            ws.send(headerBuf) 

           // this.sendMsg(ws) 测试不走ajax 发送消息
        },
        sendMsg : function (ws) {
            var self = this
            var token = '{"mid":13000000000, "shop_id":1645755332, "type":"text", "msg":"++++++++++++++测试++++++++++++++++++", "room_id":"1645755332", "dateline":1645776553, "id":"1645776553444263000"}' 
            var headerBuf = new ArrayBuffer(rawHeaderLen) 
            var headerView = new DataView(headerBuf, 0)
            var bodyBuf = self.textEncoder.encode(token) //接收一个String类型的参数返回一个Unit8Array 1个字节
            headerView.setInt32(packetOffset, rawHeaderLen + bodyBuf.byteLength) //包长度  写入从内存的第0个字节序开始  值为16 + token长度
            headerView.setInt16(headerOffset, rawHeaderLen) //写入从内存的第4个字节序开始  值为16
            headerView.setInt16(verOffset, 1) //版本号为1
            headerView.setInt32(opOffset, 4)  //写入从内存的8第个字节序开始，值为7 标识auth
            headerView.setInt32(seqOffset, 1) //从内存的12个字节序开始· 值为1   序列号（服务端返回和客户端发送一一对应）
            var flag = ws.send(self.mergeArrayBuffer(headerBuf, bodyBuf))
           
            return flag
        },

        auth: function (ws) {
            var self = this 
            //var token = '{"mid":123, "room_id":"live://1000", "platform":"web", "accepts":[1000,1001,1002]}'
            var token =  JSON.stringify(_.config.options)
            console.log(token) 
            var headerBuf = new ArrayBuffer(rawHeaderLen) 
            var headerView = new DataView(headerBuf, 0)
            var bodyBuf = self.textEncoder.encode(token) //接收一个String类型的参数返回一个Unit8Array 1个字节
            headerView.setInt32(packetOffset, rawHeaderLen + bodyBuf.byteLength) //包长度  写入从内存的第0个字节序开始  值为16 + token长度
            headerView.setInt16(headerOffset, rawHeaderLen) //写入从内存的第4个字节序开始  值为16
            headerView.setInt16(verOffset, 1) //版本号为1
            headerView.setInt32(opOffset, 7)  //写入从内存的8第个字节序开始，值为7 标识auth
            headerView.setInt32(seqOffset, 1) //从内存的12个字节序开始· 值为1   序列号（服务端返回和客户端发送一一对应）
            var flag = ws.send(self.mergeArrayBuffer(headerBuf, bodyBuf))
            //console.log("send: auth token: " + token)
            return flag
        },
        syncMsgReceived :  function (ws, body) {
            var shows = JSON.parse(body)
            if (shows) {
                for(var i=0;i<shows.length; i++){
                    _.render.show(shows[i])
                }
            }
        },
        messageReceived :  function (ws, body) {
            var self = this
            var key = _.config.options.key
            var room_id = _.config.options.room_id
            
            var dst = JSON.parse(body)
            if (dst instanceof Array) {
                for(var i=0;i<dst.length; i++){
                    _.render.show(dst[i])
                }
            } else{
                _.render.show(dst)
                self.messageAck(ws, key, room_id, dst.id) //上报 消息偏移
                //console.log( "--上报 消息偏移 --show.id-->" + show.id + " ---self.msgSeq-->" + self.msgSeq )  
            } 
        },
        mergeArrayBuffer : function (ab1, ab2) {
            var u81 = new Uint8Array(ab1), u82 = new Uint8Array(ab2),
                res = new Uint8Array(ab1.byteLength + ab2.byteLength)
            res.set(u81, 0)
            res.set(u82, ab1.byteLength)
            return res.buffer
        },
    }
    
    _.render = {
        /**
        local data = {
            mid = mid, 
            type = type,
            msg = msg,
            roomid = roomid, 
            dateline = dateline
        }
        */
        show : function(res) { //渲染消息 
            var self = this
            var html = null  
            var msg = null
            var dateline = null
            var mid = null
            var msgtype = null
            var id = null
            if (!res.msg) {
                return console.log("参数必须要有 msg")
            }
            if (!res.dateline) {
                return console.log("参数必须要有 dateline")
            }
            if (!res.mid) {
                return  console.log("参数必须要有 mid")
            }
            if (!res.type) {
                return console.log("参数必须要有 type")
            }
            
            id = res.id
            msg = res.msg
            dateline = res.dateline
            msgtype = res.type
            mid = res.mid
            
            dateline = new Date(parseInt(dateline) * 1000).toLocaleString().replace(/年|月/g, "-").replace(/日/g, " ")  
            if (msg && msg.length > 3) {
                msg = _.face.handleface( msg ) //表情过滤
            }
            if (mid == _.config.options.mid) {
                html = self.message_me(id, _.config.options.face , _.config.options.nickname,  dateline, msg, msgtype)
            } else {
                html = self.message(id, _.config.options.shop_face, _.config.options.shop_name,  dateline, msg, msgtype)
            } 
            var messageList = $("#messageList")
            //messageList.append($(html).hide().fadeIn('slow')) 
            messageList.append(html)
            //判断长度 是否需要删除 
            if (messageList.children().length > 66) {
                messageList.children().first().remove() 
            }
            _.scrollTop()
        },
        //别人发消息给我的模板； 
        message : function (id, face, nickname, time, msg, msgtype) { 
            var str = '<div class="send" id="' +id+ '">'
            str += '    <div class="time">'+time+'</div>'
            str += '    <div class="msg">'
              str += '       <img src="' + face + '" alt="头像" />' //'+face+'
            str += '         <span style=" position: absolute;left: 1.1rem;">' + nickname + '</span>'
            
            if (msgtype == 'image'){
                str += '        <pre>' + '<img src="'+ msg +'" class="msg-img"/>' + '</pre>'
            } else {
                str += '        <pre>'+msg+'</pre>'
            } 
            str += '    </div>'
            str += '</div>' 
            return str;
        },
        //我自己发消息模板； 
        message_me : function (id, face, nickname, time, msg, msgtype){ 
            var str = '<div class="show" id="' +id+ '">'
            str += '    <div class="time">'+time+'</div>'
            str += '    <div class="msg">'
            str += '        <img src="' + face + '" alt="头像" />'//'+face+'
            str += '        <span style=" position: absolute;right: 1.1rem;">' + nickname + '</span>'
            if (msgtype == 'image'){
                str += '        <pre>' + '<img src="'+ msg +'" class="msg-img" onclick="javascript:window.location.href=\'' + msg + '\'"  />' + '</pre>'
            } else {
                str += '        <pre>'+msg+' </pre>'
            } 
            str += '    </div>'
            str += '</div>' 
           return str;
        }
    }
    
    _.scrollTop = function () {
        setTimeout(function () {
            win.scrollTo(document.body.scrollWidth,document.body.scrollHeight)
            // win.scrollTo(0, document.body.scrollHeight);
        }, 200)
    }
    
    _.alert = function(msg) {
        var domsg = $("#domsg") 
        if (domsg.length < 1) {
            domsg = document.createElement("div")
            domsg.id = "domsg"
            domsg.className = "popupWindow"
            var string = '<div class="hint"><div class="text" id="msgtext" style="padding: 20px 0;" >'+ msg +'</div>';
            string += '<div class="btnBox"><a class="btnStyle btn01" id="doconfig" style="width: 100%;" >确定</a>';
            string += '</div></div>';
            domsg.innerHTML = string           
            document.body.appendChild(domsg)
            
            $("#doconfig").click(function() {
                $("#domsg").css("display","none") 
            })
        }
        $("#domsg").css("display","block")  
        $("#msgtext").html(msg)
    }

    _.face = { 
        init : function(){
            var self = this
            self.showFace() //展示工具按钮
            self.toggleFace() //切换工具栏目 并绑定事件 
            self.clickFacePic()  //点选表情图片
            self.clickFileUpload()  //点选图片上传
        }, 
        showFace : function(){ 
            $('#console_box_left').click(function () {
                $('#face_box').toggleClass('hide')
            })
        },
        toggleFace : function() {
            $('#face_box .face_box_head li').click(function () {
                var myself = $(this)
                var i = myself.data('i') 
                $("#face_box .face_box_head li, #face_box .face_box_body div").each(function(){
                    $(this).removeClass('active') //先抹去所有的 选中状态
                })
    
                myself.addClass('active') //本次的为选中状态
                $('#face_box .face_box_body div').eq(i).addClass('active')
            })
        },
        clickFileUpload : function () {
            var inputFileDom = $('#face_box_body_pic a input')
            inputFileDom.on("change", function () {
                _.upload.on(inputFileDom) 
            })
        },
        clickFacePic : function () {
            var self = this
            $('#face_box_body_qq a').click(function () {
                self.getOneFace(this)
            })
        },
        getOneFace : function(self) {
            var myself = $(self)
            var html = '[' + myself.attr('title') + ']'
            var console_box_input = $('#console_box_input')
            console_box_input.val(console_box_input.val() + html)  
        },
        handleface : function(str) {  //表情过滤     str="[得意]ddd[gg[6 ][发呆]6]]jjjj[发呆]j"
          var newstr = str
          var i=0;  var x=0;  var arr=[];
          var fn = function(s,y,p) {
                var a=fn.arguments,l=a.length,s=a[0],p=a[l-2];
                if(s=="["){ i+=1; if (i==1) { x=p; } }
                if(s=="]"){
                    i-=1;
                    if(i==0) {
                        arr.push(str.slice(x+1,p));
                    }
                    if (i<0) {i=0;}
                }
                return s;
           }
           str.replace(/[\[\]]/g, fn);   
           if (arr.length == 0 ) {
               return newstr
            }          
            var domObj = $('#face_box_body_qq a');
            for (var m in arr) {
                if ((arr[m]).length < 1 || (arr[m]).length > 5){ continue;}  //不合法表情提前过滤

                domObj.each(function(index, element) {
                    var title = element.getAttribute("title"); 
                    if ( arr[m] == title ){
                        var ls = '[' + title + ']';  
                        newstr =  newstr.replace(ls, element.innerHTML);
                        // console.log(newstr) 
                       //  newstr = newstr.replace('id="console_box_input"',  "");//替换重要id
                    }
                })  
            }
           // console.log("--匹配替换后最终结果-->" + newstr)
            return newstr
        }
    }
    
    _.upload = {
        init : function (){
            var self = this 
            // 剪切板
            document.addEventListener('paste', function(e){
                if ( !(e.clipboardData && e.clipboardData.items) ) {
                    return;
                }
                for (var i = 0, len = e.clipboardData.items.length; i < len; i++) {
                    var item = e.clipboardData.items[i];
                    //console.log(item);
                    if (item.kind === "string") {
                        item.getAsString(function (str) {
                            console.log(str);
                        })
                    } else if (item.kind === "file") {
                        var blob = item.getAsFile();
                        if (blob.size === 0) {
                            return;
                        }
                        var formdata = new FormData();
                        formdata.append("filename", blob);
                        // formdata.append('id',45);
                        _.sendAjax("http://192.168.84.168:8090/upload/file", "POST", formdata, function(result){
                            self.success(data)    
                            return 
                        }, true);
                        
                    }
                }
            })
            //剪切板 end 
        },
        inputFileDom: null,
        on : function (dom) {
            var self = this 
            self.inputFileDom = dom
            $.ajaxFileUpload({
                url : 'http://192.168.84.168:8090/upload/file?token='+_.config.options.token,
                timeout : 27000,
                headers: {
                    "token": _.config.options.token
                },
                data : {},//{'path' : $(dom).attr('path'), 'crop' : '其他字段', 'compress' : '其他字段'},
                secureuri : false,
                fileElementId : $(dom),
                dataType: 'json',
                success : function(data) {  
                    if (data['code'] == 200) { //成功  
                        self.success(data) 
                    } else {  
                        self.feild(data)   //失败
                    }
                }
            })
            return false 
        }, 
        success : function(data) {   //上传成功  
            var self = this 
            _.comet.send(data['filename'])

            $('#face_box').addClass("hide")
             
            if (self.inputFileDom && self.inputFileDom[0]) {
                var outerHTML = "上传图片" + self.inputFileDom[0]['outerHTML']
                $('#face_box_body_pic_a').html(outerHTML) 
            }
             
            _.face.clickFileUpload()  //点选图片上传
        },
        feild : function(data){
            //上传失败  
            _.alert(data.msg)
            console.log(data)
        }
    }
     
    win['_'] = _ 
    _.init()
})(window)