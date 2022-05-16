(function(win) {
    const rawHeaderLen = 16;
    const packetOffset = 0;
    const headerOffset = 4;
    const verOffset = 6;
    const opOffset = 8;
    const seqOffset = 12;


    var appendMsg = function(text) {
        var span = document.createElement("SPAN");
        var text = document.createTextNode(text);
        span.appendChild(text);
        document.getElementById("box").appendChild(span);
    }

    var Client = function(options) {
        this.options = options || {};
        var MAX_CONNECT_TIMES = 10; //最大重连次数
        var DELAY = 15000;          //每隔30秒连一次
        this.createConnect(MAX_CONNECT_TIMES, DELAY);
    }

    Client.prototype.createConnect = function(max, delay) {
        var self = this;
        if (max === 0) {
            return;
        }
        connect();

        var textDecoder = new TextDecoder();
        var textEncoder = new TextEncoder();
        var heartbeatInterval;

        function connect() {
            //var ws = new WebSocket('ws://192.168.3.222:3102/sub');
            var ws = new WebSocket(self.options.url); 
            ws.binaryType = 'arraybuffer';
            ws.onopen = function() {
                auth();
            }

            ws.onmessage = function(evt) {
                var data = evt.data;
                var dataView = new DataView(data, 0);
                var packetLen = dataView.getInt32(packetOffset);
                var headerLen = dataView.getInt16(headerOffset);
                var ver = dataView.getInt16(verOffset);
                var op = dataView.getInt32(opOffset);
                var seq = dataView.getInt32(seqOffset);

                console.log("receiveHeader: packetLen=" + packetLen, "headerLen=" + headerLen, "ver=" + ver, "op=" + op, "seq=" + seq);

                switch(op) {
                    case 8:
                        // auth reply ok
                        document.getElementById("status").innerHTML = "<color style='color:green'>auth ok<color>";
                        appendMsg("receive: auth reply");
                        // send a heartbeat to server
                        heartbeat();
                        heartbeatInterval = setInterval(heartbeat, 30 * 1000);
                        break;
                    case 3:
                        // receive a heartbeat from server
                        console.log("receive: heartbeat");
                        //appendMsg("receive: heartbeat reply");
                        break;
                    case 9:
                        // batch message
                        var offset=rawHeaderLen; 
                        for (; offset<data.byteLength; offset+=packetLen) {
                            // parse
                            var packetLen = dataView.getInt32(offset);
                            var headerLen = dataView.getInt16(offset+headerOffset);
                            var ver = dataView.getInt16(offset+verOffset);
                            var op = dataView.getInt32(offset+opOffset);
                            var seq = dataView.getInt32(offset+seqOffset);
                            var msgBody = textDecoder.decode(data.slice(offset+headerLen, offset+packetLen));
                            // callback
                            messageReceived(ver, msgBody);
                            appendMsg("receive1: ver=" + ver + " op=" + op + " seq=" + seq + " message=" + msgBody);
                        }
                        break;
                    default:
                        var msgBody = textDecoder.decode(data.slice(headerLen, packetLen));
                        messageReceived(ver, msgBody);
                        appendMsg("receive2: ver=" + ver + " op=" + op + " seq=" + seq + " message=" + msgBody);
                        break
                }
            }

            ws.onclose = function() {
                if (heartbeatInterval) clearInterval(heartbeatInterval);
                setTimeout(reConnect, delay);

                document.getElementById("status").innerHTML =  "<color style='color:red'>failed<color>";
            }

            function heartbeat() {
                var headerBuf = new ArrayBuffer(rawHeaderLen); //分配16个固定元素大小
                var headerView = new DataView(headerBuf, 0); //读写时手动设定字节序的类型
                headerView.setInt32(packetOffset, rawHeaderLen); //写入从内存的第0个字节序开始  值为16
                headerView.setInt16(headerOffset, rawHeaderLen); //写入从内存的第4个字节序开始  值为16
                headerView.setInt16(verOffset, 1);  //写入从内存的6第个字节序开始，值为1     省去的第三个参数: true为小端字节序，false为大端字节序 不填为大端字节序
                headerView.setInt32(opOffset, 2);   //写入从内存的8第个字节序开始，值为2
                headerView.setInt32(seqOffset, 1);  //写入从内存的12第个字节序开始，值为1 
                ws.send(headerBuf);
                console.log("send: heartbeat");
                //appendMsg("send: heartbeat");
            }

            function auth() {
                //协议格式对应 /api/comet/grpc/protocol
                //var token = '{"mid":123, "room_id":"live://1000", "platform":"web", "accepts":[1000,1001,1002]}'
                var token = '{"mid":' + self.options.mid + ',"key":"123456' + self.options.mid + '", "room_id":"' + self.options.room_id + '", "platform":"' + self.options.platform + '", "accepts":[' + self.options.accepts + ']}'
                var headerBuf = new ArrayBuffer(rawHeaderLen); 
                var headerView = new DataView(headerBuf, 0);
                var bodyBuf = textEncoder.encode(token); //接收一个String类型的参数返回一个Unit8Array 1个字节
                headerView.setInt32(packetOffset, rawHeaderLen + bodyBuf.byteLength); //包长度  写入从内存的第0个字节序开始  值为16 + token长度
                headerView.setInt16(headerOffset, rawHeaderLen); //写入从内存的第4个字节序开始  值为16
                headerView.setInt16(verOffset, 1); //版本号为1
                headerView.setInt32(opOffset, 7);  //写入从内存的8第个字节序开始，值为7 标识auth
                headerView.setInt32(seqOffset, 1); //从内存的12个字节序开始· 值为1   序列号（服务端返回和客户端发送一一对应）
                ws.send(mergeArrayBuffer(headerBuf, bodyBuf));

                appendMsg("send: auth token: " + token);
            }

            function messageReceived(ver, body) {
                var notify = self.options.notify;
                if(notify) notify(body);
                console.log("messageReceived:", "ver=" + ver, "body=" + body);
            }

            function mergeArrayBuffer(ab1, ab2) {
                var u81 = new Uint8Array(ab1),
                    u82 = new Uint8Array(ab2),
                    res = new Uint8Array(ab1.byteLength + ab2.byteLength);
                res.set(u81, 0);
                res.set(u82, ab1.byteLength);
                return res.buffer;
            }

            function char2ab(str) {
                var buf = new ArrayBuffer(str.length);
                var bufView = new Uint8Array(buf);
                for (var i=0; i<str.length; i++) {
                    bufView[i] = str[i];
                }
                return buf;
            }
        }

        function reConnect() {
            self.createConnect(--max, delay * 2);
        }
    }

    win['MyClient'] = Client;
})(window);
