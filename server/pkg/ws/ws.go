package ws

import (
	"encoding/json"
	"mayfly-go/pkg/logx"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024 * 1024 * 10,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var conns = make(map[uint64]*websocket.Conn, 100)

func init() {
	checkConn()
}

// 放置ws连接
func Put(userId uint64, conn *websocket.Conn) {
	existConn := conns[userId]
	if existConn != nil {
		Delete(userId)
	}

	conn.SetCloseHandler(func(code int, text string) error {
		Delete(userId)
		return nil
	})
	conns[userId] = conn
}

func checkConn() {
	heartbeat := time.Duration(60) * time.Second
	tick := time.NewTicker(heartbeat)
	go func() {
		for range tick.C {
			// 遍历所有连接，ping失败的则删除掉
			for uid, conn := range conns {
				err := conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(heartbeat/2))
				if err != nil {
					Delete(uid)
					return
				}
			}
		}
	}()
}

// 删除ws连接
func Delete(userid uint64) {
	logx.Debugf("移除websocket连接: uid = %d", userid)
	conn := conns[userid]
	if conn != nil {
		conn.Close()
		delete(conns, userid)
	}
}

// 对指定用户发送消息
func SendMsg(userId uint64, msg *Msg) {
	conn := conns[userId]
	if conn != nil {
		bytes, _ := json.Marshal(msg)
		conn.WriteMessage(websocket.TextMessage, bytes)
	}
}
