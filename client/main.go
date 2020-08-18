//// main.go
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/AlexStocks/log4go"
	"golang.org/x/net/websocket"
)

var server http.Server

func main() {
	server = http.Server{
		Addr: ":8669",
	}
	http.HandleFunc("/r", receiveHandler)                      // 接收数据
	http.Handle("/beating", websocket.Handler(handleHartBeat)) //heartbeat
	fmt.Println("*********client addr********************")
	fmt.Println("server.Addr: " + server.Addr)
	server.ListenAndServe()
}
func handleHartBeat(conn *websocket.Conn) {
	defer conn.Close()
	msg := "."
	for {
		err := websocket.Message.Send(conn, &msg)
		if err != nil {
			fmt.Println(" client heartbeat send err:" + err.Error())
			log4go.Error(err)
			conn.Close()
			break
		}
		//一秒跳一次
		time.Sleep(1 * time.Second)
	}

}
func receiveHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	// 返回时间#数据长度
	w.Write([]byte(time.Now().String() + "#" + strconv.Itoa(len(body))))

	// fmt.Println(string(body))
	// fmt.Printf("Get request result: %s\n", string(body))
	log4go.Info("完成接收 字节数bytes:" + strconv.Itoa(len(body)))
}

//func getTimeNowFormatedAsLogTime() string {
//	t := time.Now()
//	zone, _ := t.Zone()
//	tnow := t.Format("[2006/01/02 15:04:05 " + zone + "]")
//	return tnow
//}
