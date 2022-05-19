(function(win){
    var setCookie = function (name, value) {
        var date = new Date(); //获取当前时间
        var exp = 10; //expiresDays缩写exp(有效时间)
        date.setTime(date.getTime() + exp * 24 * 3600 * 1000); //格式化为cookie识别的时间
        document.cookie=escape(name) + "=" + escape(value) + ";expires="+date.toGMTString(); //将name设置为10天后过期,超过这个时间name这条cookie会消失
    }

    var getCookie = function (name) {
        var arr,reg=new RegExp("(^| )"+name+"=([^;]*)(;|$)");
        if(arr=document.cookie.match(reg)) {
            return unescape(arr[2])
        } else {
            return null
        }
    }         
    var randomNum = function (minNum,maxNum){ 
        switch(arguments.length){ 
            case 1: 
                return parseInt(Math.random()*minNum+1,10);  
            case 2: 
                return parseInt(Math.random()*(maxNum-minNum+1)+minNum,10);  
            default: 
                return 0;  
        }
    };

    var _GET = function (name){
        var reg = new RegExp("(^|&)"+ name +"=([^&]*)(&|$)");
        var r = win.location.search.substr(1).match(reg);
        if(r!=null)return  unescape(r[2]); return null;
    }
    var _ = {} 
    _.init = function() { 
        _.userlist.init() 
    } 
   
    _.userlist = {
        opt : {
            shop_id : 0,
        },
        init:function (){
            var self = this
           
            var tmp_shop_id = getCookie('shop_id')
            if (tmp_shop_id) {
                self.opt.shop_id = tmp_shop_id
            } else {
                var shop_id = _GET("shop_id") || 8000
                self.opt.shop_id = shop_id
                setCookie('shop_id', shop_id)
            }
            
            self.send(self.opt.shop_id, 0)
            setInterval(function() {
                var num = randomNum(1, 10)
                if (num % 5 == 0) {
                    self.send(self.opt.shop_id, 0)
                } else {
                    self.send(self.opt.shop_id, 1)
                }
            }, 5 * 1000)

            return true
        }, 
        send : function(shop_id, flag) {
            var self = this
            var url = "/open/finduserlist?shop_id=" + shop_id + "&flag=" + flag
            $.ajax({
                type : "POST",
                url : url,
                global: true, //希望产生全局的事件
                data : {shop_id:shop_id},
                timeout : 27000,
                cache:false,
                async: true, //是否异步
                contentType:'application/json',
                dataType:"json", // 数据格式指定为jsonp 
                //jsonp:"callback", //传递给请求处理程序或页面的，用以获得jsonp回调函数名的参数名默认就是callback
                //jsonpCallback:"getName",   // 指定回调方法
                beforeSend:function(){
                    //return console.log('发送中...')
                },
                success: function(data) {
                    if (data.success) {
                        self.opt.shop_id = shop_id 
                        if (!data.user_list || data.user_list.length <= 0) { return }
                        var len = data.user_list.length 
                        for (var i = 0; i < len; i++) { 
                            //过滤掉自己
                            self.show( data.user_list[i] ,flag)
                        }
                    } else {
                         alert("改操作需要先登录  账号密码为 8000    111111")  
                         location.href = "/login.html";
                    }
                    return 
                },
                error:function (res, status, errors){
                    console.log(res)
                    console.log(status)
                    console.log(errors)
                    console.log('消息发送失败/(ㄒoㄒ)/~~')
                    _.alert('消息发送失败/(ㄒoㄒ)/~~')
                    return
                },
                complete:function(){
                    //console.log('发送成功')
                    return 
                }
            })
            return true
        },
        show : function(m, flag) {
            var self = this
            var last = null
            if (m.last_message.length > 0) {
                last = JSON.parse(m.last_message[0])
            }
        
            // 先删除旧的
            var midLi = $("#" + m.device_id)
            if (midLi) {  midLi.remove() }
            var url = "/admin/im.html?room_id=" + m.room_id
            url += "&device_id=md5_platform_user_id" + m.shop_id
            url += "&user_id=" + m.shop_id
            url += "&nickname=" + m.shop_name
            url += "&shop_id=" + m.user_id
            url += "&shop_name=" + m.nickname
            url += "&shop_face=" + m.face
            url += "&face= " + m.shop_face
            url += "&pushurl=" + m.pushurl
            url += "&platform=" + m.platform
            url += "&suburl=" + m.suburl
            
            var html = ' <li class="mui-table-view-cell mui-media" id="'+m.device_id+'">'
            html += '<a href="'+ url +'">'
            html += '<img class="mui-media-object mui-pull-left" src="'+m.face+'" />'
            html += '<div class="mui-media-body">'
            if (m.unread > 0) {
                html += '<span style="color:red;">有新消息</span>'
            } else {
                html += '<span></span>'
            }
            html += '   ' + decodeURI(m.nickname)
            html += '        <span class="time">'+ self.format( last.dateline) +  '</span>'
            if (last && last.msg) {              
               html += '    <p class="mui-ellipsis">'+ last.msg +'.</p>'
            } else {
               html += '    <p class="mui-ellipsis">.</p>'
            }
            html += '</div>'
            if (m.num > 0) {
                html += '    <span class="mui-badge mui-badge-danger">'+m.num+'</span>'
            } 
            html += '</a>'
            html += '</li>' 
            var messageList = $("#chatlist")
            if (messageList && flag == 0) {
                if (m.online) { 
                    $(html).prependTo("#chatlist")
                } else {
                    messageList.append(html)
                }
            }
            
            if (messageList && flag == 1) {
                $(html).prependTo("#chatlist")
            }
        },
        format : function(datetime) {
            var date 
            if (datetime) {
                date = new Date(parseInt(datetime*1000));//时间戳为10位需*1000，时间戳为13位的话不需乘1000
            }  else {
               date = new Date()
            }
            var year = date.getFullYear(),
                month = ("0" + (date.getMonth() + 1)).slice(-2),
                sdate = ("0" + date.getDate()).slice(-2),
                hour = ("0" + date.getHours()).slice(-2),
                minute = ("0" + date.getMinutes()).slice(-2),
                second = ("0" + date.getSeconds()).slice(-2);
                // 拼接
            // var result = year + "-"+ month +"-"+ sdate +" "+ hour +":"+ minute +":" + second;
            var result = hour +":"+ minute +":" + second 
            return result;
        }
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
    win['_'] = _ 
    _.init()
})(window)
