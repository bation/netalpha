package main

import (
	"encoding/json"
	"fmt"
	"github.com/AlexStocks/log4go"
	ethr "github.com/ethrToPkg"
	"net/http"
	"strings"
)

func httpEthr(w http.ResponseWriter, r *http.Request) {
	AllowCrossOrigin(w)
	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}

	/// 以下为结果返回
	type result struct {
		Data []ethr.BindWithStruct `json:"data"`
		Avg  float64               `json:"avg"`
	}
	fmt.Printf("start server")
	r.ParseForm()
	ip := r.PostForm.Get("ip")
	proc := r.PostForm.Get("protocol")
	if ip == "" {
		fmt.Println("ip empty")
		return
	}
	if proc == "" {
		proc = "http"
	}
	defer r.Body.Close()
	//获取带宽
	fmt.Println("开始测试带宽")
	ethr.EthrRun(ip, proc, "b") // b bandwidth c connections/s p packets/s l 延迟
	//time.Sleep(5*time.Second)
	var ret result
	var res []ethr.BindWithStruct
	for _, val := range ethr.ResData.BandwidthArr {
		if strings.ToLower(val.Protocol) == strings.ToLower(proc) {
			res = append(res, val)
		}
		fmt.Println(val)
	}
	var sumF []float64
	for _, val := range res {
		sumF = append(sumF, val.MBits)
	}
	minVal := Minimum1(sumF)
	var tm float64
	for _, val := range res {
		if val.MBits > minVal {
			tm += val.MBits
		}
	}
	ret.Avg = tm / float64(len(res)-1)
	ret.Data = res

	ret_json, err := json.Marshal(ret)
	if err != nil {
		log4go.Error("res convert to json failed " + err.Error())
		return
	}
	w.Write([]byte(ret_json))
	ethr.ResData.Init()
	return
	fmt.Println("结束测试带宽")

}
