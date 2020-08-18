package main

import (
	"bytes"
	"fmt"
	"github.com/AlexStocks/log4go"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"io"
	"mime/multipart"
)

type addfeature struct {
	subid int    `json:"subid"`
	file  []byte `json:"file"`
}

func httpPostToTarget(addr string) string {
	// 创建表单文件
	// CreateFormFile 用来创建表单，第一个参数是字段名，第二个参数是文件名
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	writer.WriteField("sublib", "1")
	formFile, err := writer.CreateFormFile("file", "5.png")
	if err != nil {
		fmt.Println("Create form file failed: %s\n", err)
	}

	// 从文件读取数据，写入表单
	srcFile, err := os.Open("./resource/pic-03.png")
	if err != nil {
		fmt.Println("%Open source file failed: s\n", err)
	}
	defer srcFile.Close()
	_, err = io.Copy(formFile, srcFile)
	if err != nil {
		fmt.Println("Write to form file falied: %s\n", err)
	}
	// 发送表单
	contentType := writer.FormDataContentType()
	writer.Close() // 发送之前必须调用Close()以写入结尾行
	fmt.Println(getTimeNowFormatedAsLogTime())
	beforePost := time.Now()
	resp, err := http.Post(addr, contentType, buf)
	if err != nil {
		log4go.Error("传输失败：" + err.Error())
		return "{\"status\":\"fail\"}"

	}
	defer resp.Body.Close()
	var buffer [512]byte
	result := bytes.NewBuffer(nil)
	for {
		n, err := resp.Body.Read(buffer[0:])
		result.Write(buffer[0:n])
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
	}
	afterPost := time.Now()
	fmt.Println("发送时间：" + beforePost.String())
	//fmt.Println("返回值(接收时间)："+result.String())
	//strMilSec := strconv.FormatInt(afterPost.Sub(beforePost).Milliseconds(), 10)
	//数据量
	transBytes, err := strconv.ParseFloat(strings.Split(result.String(), "#")[1], 64)
	//transBytes, err := strconv.Atoi(strings.Split(result.String(),"#")[1])
	//耗时 秒
	s1 := strconv.FormatFloat(afterPost.Sub(beforePost).Seconds(), 'f', -1, 64)
	//println(strMilSec +"毫秒")
	// 速率 mb/s
	transSpeed := (transBytes / 1024 / 1024) / afterPost.Sub(beforePost).Seconds()

	println(s1 + "秒")
	println(float64ToStr(transSpeed) + "MB/s")

	return "{\"status\":\"done\",\"filesizebytes\":\"" + float64ToStr(transBytes) + "\",\"speed\":\"" + float64ToStr(transSpeed) + "\",\"takeSecond\":\"" + s1 + "\"}"

}
