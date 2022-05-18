package apihttp

import (
	"context"
	"fmt"
	"goim-demo/pkg/pb"
	"goim-demo/pkg/protocol"
	"goim-demo/pkg/util"
	"net/http"
	"strings"
	"time"
)

// apiClearData 数据清理
func (h *Router) apiClearData(w http.ResponseWriter, r *http.Request) {
	svc.ClearData(context.TODO())
	OutJson(w, OutData{Code: 200, Success: true})
}

// apiPush 数据推送
func (h *Router) apiPush(w http.ResponseWriter, r *http.Request) {
	var (
		userId string
		roomId string
		shopId string
		typ    string
		msg    string
		msgId  int64
	)

	if r.Method == "POST" {
		roomId = r.FormValue("room_id")
		typ = r.FormValue("type")
		msg = r.FormValue("msg")
		msg = strings.Replace(msg, "\r\n", "\\r\\n", -1)
		msg = strings.Replace(msg, "\r", "\\r", -1)
		msg = strings.Replace(msg, "\n", "\\n", -1)

		userId = r.FormValue("user_id")
		shopId = r.FormValue("shop_id")
	}
	if roomId == "" || typ == "" || msg == "" || userId == "" || shopId == "" {
		OutJson(w, OutData{Code: -1, Success: false, Msg: "参数room_id type msg user_id shop_id不能为空"})
		return
	}
	msgId = time.Now().UnixNano() // 消息唯一id 为了方便临时demo采用该方案， 后期线上可以用雪花算法
	body := fmt.Sprintf(`{"user_id":%s, "shop_id":%s, "type":"%s", "msg":"%s", "room_id":"%s", "dateline":%d, "id":"%d"}`,
		userId, shopId, typ, msg, roomId, time.Now().Unix(), msgId)

	buf := &pb.PushMsg{
		Type:      pb.PushMsg_ROOM,
		Operation: protocol.OpSendMsgReply,
		Speed:     2,
		Server:    config.Connect.LocalAddr,
		RoomId:    roomId,
		Msg:       util.S2B(body),
	}

	err := svc.SendRoom(context.TODO(), buf)
	if err == nil {
		// 消息持久化
		err = svc.AddMessageList(roomId, msgId, body)
	}
	if err != nil {
		OutJson(w, OutData{Code: -1, Success: false, Msg: err.Error()})
		return
	}
	OutJson(w, OutData{Code: 200, Success: true, Msg: "success"})
}
