package main

import (
	"bufio"
	"fmt"
	"github.com/AlexStocks/log4go"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"io"
)

/**
time string like "2006/01/02 15:04:05"
*/
func getHistroy(start string, end string) string {
	//stime ,_ := time.ParseDuration(start)//time.ParseInLocation("2006/01/02 15:04:05", start, time.Local)
	//etime,_ :=time.ParseDuration(end)//time.ParseInLocation("2006/01/02 15:04:05", end, time.Local)
	st, err := strconv.ParseInt(start[0:10], 10, 64)
	ed, erree := strconv.ParseInt(end[0:10], 10, 64)
	if err != nil && erree != nil {
		log4go.Error("时间戳转换失败：" + err.Error() + erree.Error())
		return ""
	}
	//转化所需模板
	timeLayout := "2006/01/02 15:04:05"

	//进行格式化
	dt1 := time.Unix(st, 0).Format(timeLayout)
	dt2 := time.Unix(ed, 0).Format(timeLayout)
	stime, _ := time.ParseInLocation(timeLayout, dt1, time.Local)
	etime, _ := time.ParseInLocation(timeLayout, dt2, time.Local)
	logFiles := listAllFileByName(1, "./log")
	var result string
	for _, val := range logFiles {
		result += getHistoryFromLogFile(val, stime, etime)
	}

	if len(result) == 0 {
		//fmt.Println("返回结果："+result)
		return result
	}
	result = "[" + result[0:len(result)-1] + "]"
	//fmt.Println("返回结果："+result)
	return result[0 : len(result)-1]

}

const UploadMessage = "upload"

// exclude 不包含的string
func getHistoryFromLogFile(path string, stime time.Time, etime time.Time) string {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		log4go.Error("open file err:" + err.Error())
	}
	reader := bufio.NewReader(file)
	result := ""
	for {
		line, err := readTheFuckingLine(reader)
		if err != nil {
			if err == io.EOF {
				// 文件末尾
				break
			} else {
				fmt.Println("file read err :")
				fmt.Println(err)
				log4go.Error(err)
				break
			}
		} else {
			// 处理读取的每行数据
			//[2020/08/03 09:41:27 CST]

			strs := strings.Split(line, "#")
			if len(strs) <= 1 {
				continue
			}
			tstr := strs[0]
			msg := strs[1]
			if !strings.Contains(msg, "{") {
				continue
			}
			if strings.Contains(msg, UploadMessage) {
				continue
			}
			t, errt := time.ParseInLocation("2006/01/02 15:04:05 CST", tstr[1:len(tstr)-1], time.Local)
			if errt != nil {
				fmt.Println(errt.Error() + ":::::asdfasdfljalsdfj")
			}
			if t.After(stime) && t.Before(etime) {
				// 时间在给定区间内
				result += msg + ","
			}
		}
		//fmt.Println("a")
	}
	return result
}

// level 目录级别 单目录1
func listAllFileByName(level int, fileDir string) []string {
	logFiles := []string{}
	pathSeparator := string(os.PathSeparator)
	var listFilePrefix string = "  "
	files, dirErr := ioutil.ReadDir(fileDir)
	if dirErr != nil {
		log4go.Error("读取目录失败" + fileDir)
	}
	tmpPrefix := ""
	for i := 1; i < level; i++ {
		tmpPrefix = tmpPrefix + listFilePrefix
	}

	for _, onefile := range files {
		if onefile.IsDir() {
			//fmt.Printf("\033[34m %s %s \033[0m \n" , tmpPrefix, onefile.Name());
			//fmt.Println(tmpPrefix, onefile.Name(), "目录:")
			listAllFileByName(level+1, fileDir+pathSeparator+onefile.Name())
		} else {
			//fmt.Print(onefile.Name()+" ")
			if strings.Contains(onefile.Name(), ".log") {
				logFiles = append(logFiles, fileDir+pathSeparator+onefile.Name())
			}
		}
	}
	return logFiles

}

//type Data struct{
//	Name string
//	Age int
//}
//type Ret struct{
//	Code int
//	Param string
//	Msg string
//	Data []Data
//}
//func HelloServer(w http.ResponseWriter, req *http.Request) {
//	data := Data{Name: "why", Age: 18}
//
//	ret := new(Ret)
//	id := req.FormValue("id")
//	//id := req.PostFormValue('id')
//
//	ret.Code = 0
//	ret.Param = id
//	ret.Msg = "success"
//	ret.Data = append(ret.Data, data)
//	ret.Data = append(ret.Data, data)
//	ret.Data = append(ret.Data, data)
//	ret_json,_ := json.Marshal(ret)
//
//	io.WriteString(w, string(ret_json))
//}
//func HelloServer1(w http.ResponseWriter, req *http.Request) {
//	io.WriteString(w, "hello, world1!\n")
//}
