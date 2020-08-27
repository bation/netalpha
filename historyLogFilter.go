package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/AlexStocks/log4go"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"io"
)

func formatDateTime(start string, end string) (time.Time, time.Time) {
	timeLayout := "2006/01/02 15:04:05"
	var stime time.Time
	var etime time.Time
	timeStr := strings.Split(start, "-")
	st := int64(0)
	ed := int64(0)
	if len(timeStr) >= 3 {
		timeLayout = "2006-01-02 15:04:05"
	} else if len(strings.Split(start, "/")) == 1 {
		_, err := strconv.ParseInt(start[0:10], 10, 64)
		if err == nil {
			st, _ = strconv.ParseInt(start[0:10], 10, 64)
			ed, _ = strconv.ParseInt(end[0:10], 10, 64)
		}
	}
	if st != 0 && ed != 0 {
		//进行格式化
		dt1 := time.Unix(st, 0).Format(timeLayout)
		dt2 := time.Unix(ed, 0).Format(timeLayout)
		stime, _ = time.ParseInLocation(timeLayout, dt1, time.Local)
		etime, _ = time.ParseInLocation(timeLayout, dt2, time.Local)
	} else {
		stime, _ = time.ParseInLocation(timeLayout, start, time.Local)
		etime, _ = time.ParseInLocation(timeLayout, end, time.Local)
	}
	return stime, etime
}

/**
time string like "2006/01/02 15:04:05" 也可以接收到秒的时间戳（10位）
status : LOSTRATEG1 OFFLINE ONLINE HIGH_LATENCY
*/
func getHistroy(start string, end string, ip string, status string) string {
	//stime ,_ := time.ParseDuration(start)//time.ParseInLocation("2006/01/02 15:04:05", start, time.Local)
	//etime,_ :=time.ParseDuration(end)//time.ParseInLocation("2006/01/02 15:04:05", end, time.Local)
	//转化所需模板
	if start == "null" || end == "null" {
		return ""
	}
	stime, etime := formatDateTime(start, end)
	// 间隔时间不能超过20min
	if etime.Sub(stime).Minutes() > 10 {
		etime = getLogEndTime("10m", stime)
	}
	logFiles := listAllFileByName(1, "./log")
	var result string
	for _, val := range logFiles {
		if !strings.Contains(val, "neta.log") {
			continue
		}
		result += getHistoryFromLogFile(val, stime, etime, strings.TrimSpace(ip), strings.TrimSpace(status))
	}

	if len(result) == 0 {
		//fmt.Println("返回结果："+result)
		return result
	}
	result = "[" + result[0:len(result)-1] + "]"
	//fmt.Println("返回结果："+result)
	return result

}
func getHistroyNetUse(start string, end string, netuse string) string {
	stime, etime := formatDateTime(start, end)
	// 间隔时间不能超过20min
	if etime.Sub(stime).Minutes() > 10 {
		etime = getLogEndTime("10m", stime)
	}
	logFiles := listAllFileByName(1, "./log")
	var result string
	for _, val := range logFiles {
		if !strings.Contains(val, "device.log") {
			continue
		}
		result += getHistoryNetUseFromLogFile(val, stime, etime, strToFloat64(netuse))
	}

	if len(result) == 0 {
		//fmt.Println("返回结果："+result)
		return result
	}
	result = "[" + result[0:len(result)-1] + "]"
	//fmt.Println("返回结果："+result)
	return result
}

const UploadMessage = "status"

// exclude 不包含的string
func getHistoryFromLogFile(path string, stime time.Time, etime time.Time, ip string, status string) string {
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
			if !(strings.HasPrefix(msg, "{") && strings.HasSuffix(msg, "}")) {
				continue
			}
			if !strings.Contains(msg, UploadMessage) {
				continue
			}
			if ip != "" {
				if !strings.Contains(msg, ip) {
					continue
				}
			}
			if status != "" {
				if !strings.Contains(msg, status) {
					continue
				}
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

// 本地网卡历史记录筛选
func getHistoryNetUseFromLogFile(path string, stime time.Time, etime time.Time, netuse float64) string {
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
			if !(strings.HasPrefix(msg, "{") && strings.HasSuffix(msg, "}")) {
				continue
			}
			if !strings.Contains(msg, "netuse") {
				continue
			}
			if netuse > 0 {
				var ds deviceSpeed
				err := json.Unmarshal([]byte(msg), &ds)
				if err != nil {
					log4go.Error("Unmarshal 转换失败 " + err.Error())
					return ""
				}
				if netuse > float64(ds.NetUse) {
					continue
				}
			}
			t, errt := time.ParseInLocation("2006/01/02 15:04:05 CST", tstr[1:len(tstr)-1], time.Local)
			if errt != nil {
				fmt.Println(errt.Error() + ":::::asdfasdfljalsdfj")
			}
			if t.After(stime) && t.Before(etime) {
				tStr := tstr[1 : len(tstr)-5]
				// fmt.Println(tStr)
				msg := msg[0:len(msg)-1] + ",\"time\":\"" + tStr + "\"}"
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
