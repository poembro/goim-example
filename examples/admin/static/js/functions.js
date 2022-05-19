function getBaseUrl() {
    var ishttps = 'https:' == document.location.protocol ? true : false;
    var url = window.location.host;
    if (ishttps) {
        url = 'https://' + url;
    } else {
        url = 'http://' + url;
    }
    return url;
}

function getWsBaseUrl() {
    var ishttps = 'https:' == document.location.protocol ? true : false;
    var url = window.location.host;
    if (ishttps) {
        url = 'wss://' + url;
    } else {
        url = 'ws://' + url;
    }
    return url;
}

function notify(title, options, callback) {
    // 先检查浏览器是否支持
    if (!window.Notification) {
        console.log("浏览器不支持notify");
        return;
    }
    var notification;
    // 检查用户曾经是否同意接受通知
    if (Notification.permission === 'granted') {
        notification = new Notification(title, options); // 显示通知
        console.log("已经获取浏览器notify权限");
    } else {
        Notification.requestPermission();
        console.log("请求浏览器notify权限");
    }
    if (notification && callback) {
        notification.onclick = function(event) {
            callback(notification, event);
        }
        setTimeout(function() {
            notification.close();
        }, 3000);
    }
}
var titleTimer = 0;
var titleNum = 0;
var originTitle = document.title;

function flashTitle() {
    if (titleTimer != 0) {
        return;
    }
    titleTimer = setInterval(function() {
        titleNum++;
        if (titleNum == 3) {
            titleNum = 1;
        }
        if (titleNum == 1) {
            document.title = '【☆】' + originTitle;
        }
        if (titleNum == 2) {
            document.title = '【★】' + originTitle;
        }
    }, 500);

}

function clearFlashTitle() {
    clearInterval(titleTimer);
    document.title = originTitle;
}
var faceTitles = ["[a]", "[b]", "[c]", "[d]", "[e]", "[f]", "[g]", "[h]", "[i]", "[j]", "[k]", "[l]", "[m]", "[n]", "[o]", "[p]", "[q]", "[r]", "[s]", "[t]", "[u]", "[v]", "[w]", "[x]", "[y]", "[z]", "[aa]", "[bb]", "[cc]", "[dd]", "[ee]", "[ff]", "[gg]", "[hh]", "[ii]", "[jj]", "[kk]", "[ll]", "[mm]", "[nn]", "[oo]", "[pp]", "[qq]", "[rr]", "[ss]", "[tt]", "[uu]", "[vv]", "[ww]", "[xx]", "[yy]", "[zz]", "[a1]", "[b1]", "[good]", "[NO]", "[c1]", "[d1]", "[e1]", "[f1]", "[g1]", "[h1]", "[i1]", "[g1]", "[k1]", "[l1]", "[m1]", "[n1]", "[o1]", "[p1]", "[q1]", "[cake]"];

function placeFace() {
    var faces = [];
    for (var i = 0; i < faceTitles.length; i++) {
        faces[faceTitles[i]] = "/static/images/face/" + i + ".gif";
    }
    return faces;
}

function replaceContent(content, baseUrl) { // 转义聊天内容中的特殊字符
    if (typeof baseUrl == "undefined") {
        baseUrl = "";
    }
    // var html = function (end) {
    //     return new RegExp('\\n*\\[' + (end || '') + '(pre|div|span|p|table|thead|th|tbody|tr|td|ul|li|ol|li|dl|dt|dd|h2|h3|h4|h5)([\\s\\S]*?)\\]\\n*', 'g');
    // };
    content = (content || '').replace(/&(?!#?[a-zA-Z0-9]+;)/g, '&amp;')
        .replace(/<(?!br).*?>/g, '') // 去掉html
        .replace(/\\n/g, '<br>') // 转义换行
    content = replaceSpecialTag(content, baseUrl);
    return content;
}

function replaceSpecialTag(str, baseUrl) {
    var faces = placeFace();
    if (typeof baseUrl == "undefined") {
        baseUrl = "";
    }
    str = str.replace(/face\[([^\s\[\]]+?)\]/g, function(face) { // 转义表情
            var alt = face.replace(/^face/g, '');
            return '<img alt="' + alt + '" title="' + alt + '" src="' + baseUrl + faces[alt] + '">';
        })
        .replace(/img\[([^\s\[\]]+?)\]/g, function(face) { // 转义图片
            var src = face.replace(/^img\[/g, '').replace(/\]/g, '');;
            return '<a href="' + baseUrl + src + '" target="_blank"><img data-src="' + baseUrl + src + '" data-lightbox class="chatImagePic"  src="' + baseUrl + src + '?width=400"/></a>';
        })
        .replace(/audio\[([^\s\[\]]+?)\]/g, function(face) { // 转义图片
            var src = face.replace(/^audio\[/g, '').replace(/\]/g, '');;
            return '<div class="chatAudio"><audio controls ref="audio" src="' + src + '" class="audio"></audio></div>';
        })
        .replace(/file\[([^\s\[\]]+?)\]/g, function(face) { // 转义图片
            var src = face.replace(/^file\[/g, '').replace(/\]/g, '');;
            return '<div class="folderBtn" onclick="window.open(\'' + baseUrl + src + '\')"  style="font-size:25px;"/>';
        })
        .replace(/\[([^\]]+?)\]+link\[([^\]]+?)\]/g, function(face) { // 转义超链接
            var text = face.replace(/link\[.*?\]/g, '').replace(/\[|\]/g, '');
            var src = face.replace(/^\[([^\s\[\]]+?)\]+link\[/g, '').replace(/\]/g, '');
            return '<a href="' + src + '" target="_blank"/>' + text + '</a>';
        })
        .replace(/product\[([^\[\]]+?)\]/g, function(product) {
            if (!arguments[1]) {
                return;
            }
            var jsonStr = arguments[1].replace(/\'/g, '"');
            console.log(jsonStr);
            try {
                var info = JSON.parse(jsonStr);
                if (typeof info == "undefined") {
                    return;
                }
                if (info.title) {
                    var title = info.title;
                } else {
                    var title = "GOFLY客服系统";
                }
                if (info.price) {
                    var price = info.price;
                }
                if (info.img) {
                    var img = "<img src='" + info.img + "'/>";
                } else {
                    var img = "";
                }
                if (info.url) {
                    var url = info.url;
                } else {
                    var url = "https://gofly.sopans.com";
                }
                var html = `
                    <a class="productCard" href="` + url + `" target="_blank"/>
                    ` + img + `

                    <div class="productCardTitle">
                        <p class="productCardTitle">` + title + `</p>
                        <p class="productCardPrice">` + price + `</p>
                    </div>
                    </a>
            `;
                return html;
            } catch (e) {
                return jsonStr;
            }

        });
    return str;
}

function bigPic(src, isVisitor) {
    alert(src);
    if (isVisitor) {
        window.open(src);
        return;
    }
}

function filter(obj) {
    var imgType = ["image/jpeg", "image/png", "image/jpg", "image/gif"];
    var filetypes = imgType;
    var isnext = false;
    for (var i = 0; i < filetypes.length; i++) {
        if (filetypes[i] == obj.type) {
            return true;
        }
    }
    return false;
}

function sleep(time) {
    var startTime = new Date().getTime() + parseInt(time, 10);
    while (new Date().getTime() < startTime) {}
}

function checkLang() {
    var langs = ["cn", "en", "jp"];
    var lang = getQuery("lang");
    if (lang != "" && langs.indexOf(lang) > 0) {
        return lang;
    }
    var lang = getLocalStorage("lang");
    if (lang) {
        return lang;
    }
    return "cn";
}

function changeURLPar(destiny, par, par_value) {
    var pattern = par + '=([^&]*)';
    var replaceText = par + '=' + par_value;
    if (destiny.match(pattern)) {
        var tmp = '/\\' + par + '=[^&]*/';
        tmp = destiny.replace(eval(tmp), replaceText);
        return (tmp);
    } else {
        if (destiny.match('[\?]')) {
            return destiny + '&' + replaceText;
        } else {
            return destiny + '?' + replaceText;
        }
    }
    return destiny + '\n' + par + '\n' + par_value;
}

function getQuery(key) {
    var query = window.location.search.substring(1);
    var key_values = query.split("&");
    var params = {};
    key_values.map(function(key_val) {
        var key_val_arr = key_val.split("=");
        params[key_val_arr[0]] = key_val_arr[1];
    });
    if (typeof params[key] != "undefined") {
        return params[key];
    }
    return "";
}

function utf8ToB64(str) {
    return window.btoa(unescape(encodeURIComponent(str)));
}

function b64ToUtf8(str) {
    return decodeURIComponent(escape(window.atob(str)));
}
//存储localStorge
function setLocalStorage(key, obj) {
    if (!navigator.cookieEnabled || typeof window.localStorage == 'undefined') {
        return false;
    }
    localStorage.setItem(key, JSON.stringify(obj));
    return true;
}
//读取localStorge
function getLocalStorage(key) {
    if (!navigator.cookieEnabled || typeof window.localStorage == 'undefined') {
        return false;
    }
    var str = localStorage.getItem(key);
    if (!str) {
        return false;
    }
    return JSON.parse(str);
}
var imgs = document.querySelectorAll('img');

//offsetTop是元素与offsetParent的距离，循环获取直到页面顶部
function getTop(e) {
    var T = e.offsetTop;
    while (e = e.offsetParent) {
        T += e.offsetTop;
    }
    return T;
}

function lazyLoad(imgs) {
    var H = document.documentElement.clientHeight; //获取可视区域高度
    var S = document.documentElement.scrollTop || document.body.scrollTop;
    for (var i = 0; i < imgs.length; i++) {
        if (H + S > getTop(imgs[i])) {
            console.log(imgs[i]);
            imgs[i].src = imgs[i].getAttribute('data-src');
        }
    }
}

function loadImage(url) {
    var image = new Image();
    image.src = url;
    console.log(image);
}

function image2Canvas(image) {
    var canvas = document.createElement('canvas')
    var ctx = canvas.getContext('2d')
    canvas.width = image.naturalWidth
    canvas.height = image.naturalHeight
    ctx.drawImage(image, 0, 0, canvas.width, canvas.height)
    return canvas
}

function canvas2DataUrl(canvas, quality, type) {
    return canvas.toDataURL(type || 'image/jpeg', quality || 0.8)
}

function dataUrl2Image(dataUrl, callback) {
    var image = new Image()
    image.onload = function() {
        callback(image)
    }
    image.src = dataUrl
}

function dateFormat(fmt, date) {
    let ret;
    const opt = {
        "Y+": date.getFullYear().toString(), // 年
        "m+": (date.getMonth() + 1).toString(), // 月
        "d+": date.getDate().toString(), // 日
        "H+": date.getHours().toString(), // 时
        "M+": date.getMinutes().toString(), // 分
        "S+": date.getSeconds().toString() // 秒
        // 有其他格式化字符需求可以继续添加，必须转化成字符串
    };
    for (let k in opt) {
        if (k != "Y+") {
            var length = opt[k].length;
            if (length < 2) {
                opt[k] = "0" + opt[k];
            }
        }
        ret = new RegExp("(" + k + ")").exec(fmt);
        if (ret) {
            fmt = fmt.replace(ret[1], (ret[1].length == 1) ? (opt[k]) : (opt[k].padStart(ret[1].length, "0")))
        };
    };
    return fmt;
}
/**
 * 人性化时间
 * @param {Object} timestamp
 */
function beautifyTime(timestamp, lang) {
    var mistiming = Math.round(new Date() / 1000) - timestamp;
    mistiming = Math.abs(mistiming)
    if (lang == "en") {
        var postfix = mistiming > 0 ? 'ago' : 'later'
        var arrr = [' years ', ' months ', ' weeks ', ' days ', ' hours ', ' minutes ', ' seconds '];
        var just = 'just now';
    } else {
        var postfix = mistiming > 0 ? '前' : '后'
        var arrr = ['年', '个月', '周', '天', '小时', '分钟', '秒'];
        var just = '刚刚';
    }
    if (mistiming <= 1) {
        return just;
    }

    var arrn = [31536000, 2592000, 604800, 86400, 3600, 60, 1];

    for (var i = 0; i < 7; i++) {
        var inm = Math.floor(mistiming / arrn[i])
        if (inm != 0) {
            return inm + arrr[i] + postfix
        }
    }
}

/**
 * 判断是否是手机访问
 * @returns {boolean}
 */
function isMobile() {
    if (/Mobile|Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent)) {
        return true;
    }
    return false;
}
//发送ajax的助手函数
function sendAjax(url, method, params, callback) {
    $.ajax({
        type: method,
        url: url,
        data: params,
        headers: {
            "token": localStorage.getItem("token")
        },
        success: function(data) {
            callback(data);
        }
    });
}
//复制文本
function copyText(text) {
    var target = document.createElement('input') //创建input节点
    target.value = text // 给input的value赋值
    document.body.appendChild(target) // 向页面插入input节点
    target.select() // 选中input
    document.execCommand("copy"); // 执行浏览器复制命令
    document.body.removeChild(target);
    return true;
}

function MyHereDoc() {
    /*HERE

    HERE*/
    var here = "HERE";
    var reobj = new RegExp("/\\*" + here + "\\n[\\s\\S]*?\\n" + here + "\\*/", "m");
    str = reobj.exec(MyHereDoc).toString();
    str = str.replace(new RegExp("/\\*" + here + "\\n", 'm'), '').toString();
    return str.replace(new RegExp("\\n" + here + "\\*/", 'm'), '').toString();
}
//js获取当前时间
function getNowDate() {
    var myDate = new Date;
    var year = myDate.getFullYear(); //获取当前年
    var mon = myDate.getMonth() + 1; //获取当前月
    var date = myDate.getDate(); //获取当前日
    var hours = myDate.getHours(); //获取当前小时
    var minutes = myDate.getMinutes(); //获取当前分钟
    var seconds = myDate.getSeconds(); //获取当前秒
    var now = year + "-" + mon + "-" + date + " " + hours + ":" + minutes + ":" + seconds;
    return now;
}
//获取当前时间戳
function getTimestamp() {
    return new Date(getNowDate()).getTime();
}
//删除对象中的空属性
function removePropertyOfNull(obj) {
    var i = obj.length;
    while (i--) {
        if (obj[i] === null) {
            obj.splice(i, 1);
        }
    }
    return obj;
}
//判断版本号大小
function compareVersion(v1, v2) {
    if (v1 == v2) {
        return 0;
    }

    const vs1 = v1.split(".").map(a => parseInt(a));
    const vs2 = v2.split(".").map(a => parseInt(a));

    const length = Math.min(vs1.length, vs2.length);
    for (let i = 0; i < length; i++) {
        if (vs1[i] > vs2[i]) {
            return 1;
        } else if (vs1[i] < vs2[i]) {
            return -1;
        }
    }

    if (length == vs1.length) {
        return -1;
    } else {
        return 1;
    }
}

function isWeiXin() {
    //window.navigator.userAgent属性包含了浏览器类型、版本、操作系统类型、浏览器引擎类型等信息，这个属性可以用来判断浏览器类型
    var ua = window.navigator.userAgent.toLowerCase();
    //通过正则表达式匹配ua中是否含有MicroMessenger字符串
    if (ua.match(/MicroMessenger/i) == 'micromessenger') {
        return true;
    } else {
        return false;
    }
};