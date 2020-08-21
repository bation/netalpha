package main

import (
	"encoding/json"
	"fmt"
	"github.com/AlexStocks/log4go"
	ethr "github.com/ethrToPkg"
	"net/http"
)

func httpEthr(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("start server")
	r.ParseForm()
	ip := r.PostForm.Get("ip")
	if ip == "" {
		fmt.Println("ip empty")
		return
	}
	//获取带宽
	fmt.Println("开始测试带宽")
	ethr.EthrRun(ip, "http", "b") // b bandwidth c connections/s p packets/s l 延迟
	fmt.Println("结束测试带宽")
	if ethr.FinishFlag {
		/// 以下为结果返回
		type result struct {
			Data map[int]ethr.BindWithStruct `json:"data"`
		}
		var ret result
		ret.Data = ethr.BandwidthMap

		ret_json, err := json.Marshal(ret)
		if err != nil {
			log4go.Error("res convert to json failed " + err.Error())
			return
		}
		w.Write([]byte(ret_json))
		return
	}

}
