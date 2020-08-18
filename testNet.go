package main

import (
	"fmt"
	"io"
	"net"
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

// del标记 未使用的方法
func ListenAndReceiveFile() {

	var addr = "ws://" + cfg.Ip + ":" + server.Addr + "/echo" //你的地址(IP:PORT)
	listenner, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("net.Listen err =", err)
		log4go.Error(err)
		return
	}
	defer listenner.Close()

	conn, errl := listenner.Accept()
	if errl != nil {
		fmt.Println("listenner.Accept err =", errl)
		log4go.Error(err)
		return
	}
	var n int
	buf := make([]byte, 1024)
	n, err = conn.Read(buf)
	if err != nil {
		fmt.Println("conn.Read fileName err =", err)
		log4go.Error(err)
		return
	}
	fileName := string(buf[:n])
	n, err = conn.Write([]byte("ok"))
	if err != nil {
		fmt.Println("conn.Write ok err =", err)
		log4go.Error(err)
		return
	}

	RecvFile(fileName, conn)
}

func RecvFile(fileName string, conn net.Conn) {
	erre := os.Remove(fileName)
	if erre != nil {
		// 删除失败
	} else {
		// 删除成功
	}
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println("os.Create err =", err)
		log4go.Error(err)
		return
	}

	defer file.Close()

	buf := make([]byte, 1024*4)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {

				fmt.Println("文件接收完成,时间：" + time.Now().String())
			} else {
				fmt.Println("conn.Read err =", err)
				log4go.Error(err)
			}
			return
		}

		n, err = file.Write(buf[:n])
		if err != nil {
			fmt.Println("file.Write err =", err)
			log4go.Error(err)
			break
		}
	}
}
