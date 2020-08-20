package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/AlexStocks/log4go"
	"golang.org/x/net/websocket"
)

func SendFile(path string, conn *websocket.Conn) int64 {
	file, err := os.Open(path)

	if err != nil {
		fmt.Println("os.Open err =", err)
		log4go.Error(err)
		return -1
	}
	defer file.Close()
	buf := make([]byte, 4096)
	// 开始传输
	beginTime := time.Now()
	fmt.Println("开始传输时间：" + beginTime.String())
	count := 0.0
	for {
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("文件发送完毕 io.eof")
			} else {
				fmt.Println("file.Read err =", err)
				log4go.Error(err)
			}

			break
		}
		count += float64(n)
		if n == 0 {
			fmt.Println("文件发送完毕 buf=0")
			break
		}
		websocket.Message.Send(conn, buf[0:n])
		// conn.Write(buf[:n])
	}
	//传输结束
	fmt.Printf("传输数据量：%f MB  \n", (count / 1024.0 / 1024.0))
	endTime := time.Now()
	fmt.Println("结束传输时间：" + endTime.String())
	deltSecond := endTime.Sub(beginTime)
	fmt.Println("传输用时：%f ms " + strconv.FormatInt(deltSecond.Milliseconds(), 10))
	var fltSec = deltSecond.Milliseconds()
	return fltSec
}
