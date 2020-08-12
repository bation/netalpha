package main

import (
	"encoding/json"
	"fmt"
	lgg "github.com/AlexStocks/log4go"
	"strconv"
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
		lgg.Error(err.Error())
	}
	return intnum
}
func strToInt64(str string) int64 {
	st, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		lgg.Error(err.Error())
	}
	return st

}
func strToFloat64(flo string) float64 {
	res, err := strconv.ParseFloat(flo, 64)
	if err != nil {
		lgg.Error(err.Error())
	}
	return res
}
func structToJsonsting(v interface{}) string{
	data, err := json.Marshal(v)
	if err!=nil{
		fmt.Println("structToJsonsting 失败："+err.Error())
	}
	return  string(data)
}
func jsonToStruct(msg string,stt interface{}) interface{}{
	err := json.Unmarshal([]byte(msg), &stt)
	if err!=nil{
		fmt.Println("转换失败jsontostruct:"+err.Error())
	}
	return stt
}
