package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/AlexStocks/log4go"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func float64ToStr(flt float64) string {
	return strconv.FormatFloat(flt, 'f', -1, 64)
}
func int64ToString(val int64) string {
	return strconv.FormatInt(val, 10)
}
func intToStr(num int) string {
	return strconv.Itoa(num)
}
func strToInt(num string) int {
	intnum, err := strconv.Atoi(num)
	if err != nil {
		log4go.Error(err.Error())
	}
	return intnum
}
func strToInt64(str string) int64 {
	st, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		log4go.Error(err.Error())
	}
	return st

}
func strToFloat64(flo string) float64 {
	res, err := strconv.ParseFloat(flo, 64)
	if err != nil {
		log4go.Error(err.Error())
	}
	return res
}
func structToJsonsting(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		fmt.Println("structToJsonsting 失败：" + err.Error())
	}
	return string(data)
}
func jsonToStruct(msg []byte, stt interface{}) interface{} {
	err := json.Unmarshal(msg, &stt)
	if err != nil {
		fmt.Println("转换失败jsontostruct:" + err.Error())
	}
	return stt
}

//获取URL的GET参数
func GetUrlArg(r *http.Request, name string) string {
	var arg string
	values := r.URL.Query()
	arg = values.Get(name)
	return arg
}
func GetPostArg(r *http.Request, name string) string {
	var arg string
	err := r.ParseForm() // Parses the request body
	if err != nil {
		fmt.Println("获取参数失败：" + err.Error())
	}
	arg = r.Form.Get(name) // x will be "" if parameter is not set
	return arg
}
func getLogEndTime(min string, startTime time.Time) time.Time {
	var endTime time.Time
	if !strings.Contains(min, "m") {
		remainingMin, _ := time.ParseDuration(min + "m")
		endTime = startTime.Add(remainingMin)

	} else {
		remainingMin, _ := time.ParseDuration(min)
		endTime = startTime.Add(remainingMin)
	}
	return endTime
}
func StringIpToInt(ipstring string) int {
	ipSegs := strings.Split(ipstring, ".")
	var ipInt int = 0
	var pos uint = 24
	for _, ipSeg := range ipSegs {
		tempInt, _ := strconv.Atoi(ipSeg)
		tempInt = tempInt << pos
		ipInt = ipInt | tempInt
		pos -= 8
	}
	return ipInt
}

func IpIntToString(ipInt int) string {
	ipSegs := make([]string, 4)
	var len int = len(ipSegs)
	buffer := bytes.NewBufferString("")
	for i := 0; i < len; i++ {
		tempInt := ipInt & 0xFF
		ipSegs[len-i-1] = strconv.Itoa(tempInt)
		ipInt = ipInt >> 8
	}
	for i := 0; i < len; i++ {
		buffer.WriteString(ipSegs[i])
		if i < len-1 {
			buffer.WriteString(".")
		}
	}
	return buffer.String()
}

//错误处理函数
func checkErr(err error, extra string) bool {
	if err != nil {
		formatStr := " Err : %s\n"
		if extra != "" {
			formatStr = extra + formatStr
		}

		fmt.Fprintf(os.Stderr, formatStr, err.Error())
		return true
	}

	return false
}
