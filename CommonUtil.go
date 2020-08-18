package main

import (
	"encoding/json"
	"fmt"
	"github.com/AlexStocks/log4go"
	"net/http"
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
