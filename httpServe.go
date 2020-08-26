package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/websocket"

	"github.com/AlexStocks/log4go"
)

func getTimeNowFormatedAsLogTime() string {
	t := time.Now()
	zone, _ := t.Zone()
	tnow := t.Format("[2006/01/02 15:04:05 " + zone + "]")
	return tnow
}

// 没看懂 两重for循环readline
func readTheFuckingLine(r *bufio.Reader) (string, error) {
	line, isprefix, err := r.ReadLine()
	for isprefix && err == nil {
		var bs []byte
		bs, isprefix, err = r.ReadLine()
		line = append(line, bs...)
	}
	return string(line), err
}
func handleInfo(ws *websocket.Conn) {
	defer ws.Close()
	for {
		for i := 0; i < len(cfg.Targets); i++ {
			data := statusQuene.GetData(uint(i))
			if data != nil {

				err := websocket.Message.Send(ws, data.(string))
				if err != nil {
					fmt.Println("发送失败，连接可能关闭 err")
					fmt.Println(err)
					log4go.Error(err)
					return
				}
			}
		}
		time.Sleep(time.Duration(cfg.Interval) * time.Second)
	}
}

// func handleInfo(ws *websocket.Conn) {
// 	for {
// 		info := chanelSpeedAndPingRcver.Read().(map[string]int)

// 		mjson, _ := json.Marshal(info)
// 		mString := string(mjson)
// 		sumStr := handleJsonIntToIP(mString, info["ip"])
// 		err := websocket.Message.Send(ws, sumStr)
// 		if err != nil {
// 			fmt.Println("chanelSpeedAndPingRcver err")
// 			fmt.Println(err)

// 			break
// 		}

// 		// fmt.Printf("%s 实时网速：↑ %.2fkb/s ↓ %.2fkb/s \n", IpIntToString(info["ip"]), float32(info["upload"])/1024/1, float32(info["download"])/1024/1)

// 	}
// }

// 异常节点测试 读取数据
func handleTargetSocket(ws *websocket.Conn) {
	defer ws.Close()
	rFiles := GetRunningFiles()
	// //读取log
	var ch chan int
	for _, val := range rFiles {
		go readLog("./log/"+val, ch, ws)
	}
	for i := 0; i < len(rFiles); i++ {
		<-ch
	}

}
func readLog(path string, ch chan int, ws *websocket.Conn) {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		fmt.Println("read log file err:")
		fmt.Println(err)
		log4go.Error(err)
		return
	}
	reader := bufio.NewReader(f)
	isEOFOnce := false // false 从最后一行发
	// true 从第一行发
	isEOFOnce = true
	fmt.Println("read tar log now")
	lineCount := 0
	for {

		line, err := readLine(reader)
		if err != nil {
			if err == io.EOF {
				// 文件末尾
				isEOFOnce = true // 读到了文件末尾
				// 每次到文件末尾休息5秒  因为5秒发一次
				time.Sleep(5 * time.Second)
			} else {
				fmt.Println("file read err or conn lost:")
				fmt.Println(err)
				log4go.Error(err)
				break
			}
		} else {
			// err = nil 时处理消息

			// 首次发送数据从最后一行开始发
			if isEOFOnce {
				line = strings.TrimSpace(line)
				// fmt.Println(line)
				if strings.Contains(line, "因到达运行时间而终止") {
					break
				}
				splitLine := strings.Split(line, "#")
				if len(splitLine) < 2 {
					continue
				}
				tStr := splitLine[0][1 : len(splitLine[0])-5]
				// fmt.Println(tStr)
				msgSend := splitLine[1][0:len(splitLine[1])-1] + ",\"time\":\"" + tStr + "\"}"
				// fmt.Println((msgSend))
				err := websocket.Message.Send(ws, msgSend)

				if err != nil {
					fmt.Println("发送失败，连接可能关闭 err")
					log4go.Error(err)

					break
				}
			}
			lineCount += 1 //行计数
			//if lineCount >= (50000) {
			//	fmt.Println("文件到达日志最大行，准备重新读取文件")
			//	log4go.Info("文件到达日志最大行，准备重新读取文件")
			//	ferr := f.Close() // 不然会一直占用io导致日志无法rotate
			//	if ferr != nil {
			//		fmt.Println(ferr)
			//	}
			//	time.Sleep(1 * time.Second) // 只能休眠<3秒否则icmp可能连接失败
			//	goto lableReachLineMax
			//}

		}
	}
	ch <- 1
}

func startPingTargetStandaloneFunc(w http.ResponseWriter, r *http.Request) {
	AllowCrossOrigin(w)

	// 传输的json必须是字符串格式的json string
	defer r.Body.Close()
	type RequestSocketUrl struct {
		Min     int      `json:"min"`
		Targets []string `json:"targets"`
	}
	var data RequestSocketUrl

	bd, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log4go.Error("read body failed " + err.Error())
		return
	}
	if err := json.Unmarshal(bd, &data); err != nil {
		log4go.Error("format body to json failed " + err.Error())
		return
	}
	fmt.Printf("min: %d ,tar:%s \n", data.Min, data.Targets)
	//if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
	//	fmt.Println("decode err:"+ err.Error())
	//}
	//查重，查看是否已有ping任务在运行
	if isRepeatTask(data.Targets) {
		log4go.Error("已在测试中，同时间段重复测试无效")
		ret := historyRes{Data: "已在测试中，同时间段重复测试无效"}
		w.WriteHeader(403)
		ret_json, _ := json.Marshal(ret)
		w.Write([]byte(ret_json))
		return
	}
	// 开始ping
	for _, ip := range data.Targets {
		fmt.Println(ip + "|||" + intToStr(data.Min))
		// 获得logger后，需要ping target，每个ping和logger都是独立的.
		// 修改ping，加个logger参数？
		var ipLogger = getNewLogger(ip, intToStr(data.Min))
		// 开始独立记录
		go GoPing([]string{ip}, true, &ipLogger, data.Min)

	}

	/// 以下为结果返回
	rres := "done"
	type result struct {
		Data string `json:"data"`
	}
	var ret result
	ret.Data = rres

	ret_json, err := json.Marshal(ret)
	if err != nil {
		log4go.Error("res convert to json failed " + err.Error())
		return
	}
	w.Write([]byte(ret_json))
	return
}

func startLiteServer() {
	var server = http.Server{
		Addr: ":8769",
	}
	fmt.Println("*********server addr********************")
	fmt.Println("server.Addr: " + server.Addr)
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("服务启动失败" + err.Error())
	}
}

// 网络使用情况开关
func controlNetUsing(w http.ResponseWriter, r *http.Request) {
	var stat string
	if r.Method == "POST" {
		stat = GetPostArg(r, "msg")
	}
	fmt.Println(stat)
	ret := historyRes{Data: ""}
	if strings.Contains(stat, "start") {
		// 开始网络流量监控
		ret.Data = "{\"result\":\"" + "start" + "\"}"
		//done 重复验证 只能运行一个
		if netUsingQuene.GetStatus() {
			w.WriteHeader(403)
			w.Write([]byte("同一时间只能运行一个网络监控！"))
			return
		}
		// 清空队列
		for netUsingQuene.Size() > 0 {
			netUsingQuene.Dequeue()
		}
		go DeviceSpeed()

	} else {
		// 停止网络流量监控
		ret.Data = "{\"result\":\"" + "stop" + "\"}"
		// 重复验证
		if netUsingQuene.Contains("stop") {
			w.WriteHeader(403)
			w.Write([]byte("网络监控已经停止"))
			return
		}
		// 添加一条数据就会停止运行
		netUsingQuene.Enqueue("stop")
	}
	ret_json, err := json.Marshal(ret)
	if err != nil {
		fmt.Println(err.Error())
	}
	w.Write([]byte(ret_json))

}

// 读取队列数据 net
func handleNetUsing(conn *websocket.Conn) {
	defer conn.Close()
	for {
		if !netUsingQuene.GetStatus() {
			break
		}
		if netUsingQuene.Contains("stop") {
			break
		}
		if netUsingQuene.Peek() == nil {
			continue
		}
		var msg = netUsingQuene.Peek().(string)
		//fmt.Println(msg)
		//fmt.Printf("\n")
		err := websocket.Message.Send(conn, msg)
		if err != nil {
			log4go.Error("netusinginfo 发送失败：%s", err.Error())
			break
		}
		time.Sleep(2 * time.Second)

	}

}

// 反馈节点掉线，
func reportNodeDownFunc(ip string, downtime string) []byte {
	//done 反馈节点掉线接口
	if cfg.OfflineRepURL == "" {
		return nil
	}
	var ptc PostMapData
	ptc.Url = cfg.OfflineRepURL
	ptc.Data["ip"] = ip
	ptc.Data["downtime"] = downtime
	resp, err := ptc.PostWithFormData()
	if err != nil {
		log4go.Error("反馈失败：" + err.Error())
	}
	return resp

}

// 获取网卡流量历史记录
func getHistroyNetUseFunc(w http.ResponseWriter, r *http.Request) {
	AllowCrossOrigin(w)

	var d1 string
	var d2 string
	var netuse string
	if r.Method == "POST" {
		d1 = GetPostArg(r, "d1")
		d2 = GetPostArg(r, "d2")
		netuse = GetPostArg(r, "netuse")
	}
	if strings.TrimSpace(netuse) == "" {
		netuse = "0"
	}
	fmt.Println(d1 + "|||" + d2)
	res := getHistroyNetUse(d1, d2, netuse)
	ret := historyRes{Data: ""}
	ret.Data = "{\"result\":" + res + "}"
	//fmt.Println(ret.Result)
	if len(res) > 100 {
		fmt.Println(ret.Data[0:100] + "……")
	}
	if res == "" {
		ret.Data = "{\"result\":\"" + res + "\"}"
		fmt.Println(ret.Data)
	}
	ret_json, err := json.Marshal(ret)
	if err != nil {
		fmt.Println(err.Error())
	}
	w.Write([]byte(ret_json))
}

// 获取通断历史记录
func getHistoryFunc(w http.ResponseWriter, r *http.Request) {
	AllowCrossOrigin(w)
	var d1 string
	var d2 string
	var ip string
	var status string
	if r.Method == "POST" {
		d1 = GetPostArg(r, "d1")
		d2 = GetPostArg(r, "d2")
		ip = GetPostArg(r, "ip")
		status = strings.ToUpper(GetPostArg(r, "status"))
	}
	fmt.Println(d1 + "|||" + d2)
	//body, _ := ioutil.ReadAll(r.Body)
	//fmt.Println(string(body))
	//data := strings.Split(string(body), "&")
	//ds1 := strings.Split(data[0], "=")
	//ds2 := strings.Split(data[1], "=")
	//d1 := ds1[1]
	//d2 := ds2[1]
	res := getHistroy(d1, d2, ip, status)
	ret := historyRes{Data: ""}
	ret.Data = "{\"result\":" + res + "}"
	//fmt.Println(ret.Result)
	if len(res) > 100 {
		fmt.Println(ret.Data[0:100] + "……")
	}
	if res == "" {
		ret.Data = "{\"result\":\"" + res + "\"}"
		fmt.Println(ret.Data)
	}
	ret_json, err := json.Marshal(ret)
	if err != nil {
		fmt.Println(err.Error())
	}
	w.Write([]byte(ret_json))
	// 返回时间#数据长度
	//w.Write([]byte(time.Now().String() + "#" + strconv.Itoa(len(body))))
}

// 更新配置文件
func updateConfig(w http.ResponseWriter, r *http.Request) {
	AllowCrossOrigin(w)
	defer r.Body.Close()
	r.ParseForm()
	var jsonMap url.Values
	jsonMap = r.PostForm
	ret := historyRes{Data: ""}
	ret.Data = "ok"
	ret_json, err := json.Marshal(ret)
	if err != nil {
		fmt.Println(err.Error())
	}
	if isok, _ := cfg.SetValues(jsonMap); isok {
		w.Write([]byte(ret_json))
	}
	//fmt.Printf("jsonMap:%s \n",jsonMap)
}
func httpHandle() {
	http.HandleFunc("/testNetFlow", httpEthr) // 传输数据 ethr
	// http.Handle("/css/", http.FileServer(http.Dir("template")))
	//http.HandleFunc("/echo", echoFunc)                             // 传输数据 读写
	//http.Handle("/config", websocket.Handler(handleConfigRequest)) // 配置 读写
	//http.HandleFunc("/sendRestart", restartProgram)     //只读
	//http.HandleFunc("/admin", adminTemplateFunc)
	http.HandleFunc("/startPingTargets", startPingTargetStandaloneFunc)       // 开始ping选择的地址
	http.Handle("/getPingTargetsInfo", websocket.Handler(handleTargetSocket)) // 获取ping 选择的地址的日志
	//http.HandleFunc("/controlNetUsing", controlNetUsing)                      // 控制网络流量开关
	http.Handle("/getNetUsingInfo", websocket.Handler(handleNetUsing)) //获取网络流量websocket
	http.Handle("/js/", http.FileServer(http.Dir("template")))         // 文件服务
	http.Handle("/info", websocket.Handler(handleInfo))                // 获取通断信息
	http.HandleFunc("/getHistory", getHistoryFunc)                     // 获取通断历史记录接口
	http.HandleFunc("/getHistoryNetUse", getHistroyNetUseFunc)         // 获取网卡流量历史记录接口
	http.HandleFunc("/updateConfig", updateConfig)                     //修改配置文件
	// 以上接口 ****** 以下路由
	http.HandleFunc("/", indexFunc)                  //入口
	http.HandleFunc("/node", templateNode)           //入口
	http.HandleFunc("/netflow", templateNetFlow)     //入口
	http.HandleFunc("/exception", templateException) //入口
	http.HandleFunc("/config", templateConfig)       //入口

}

func templateNode(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./template/node.html")
	if err != nil {
		fmt.Fprintln(w, err)
		log4go.Error(err)
		return
	}
	var runningTargets []string
	files := GetRunningFiles()
	for _, val := range files {
		valIp := strings.Split(val, "_")[0]
		runningTargets = append(runningTargets, valIp)
	}
	cfg.RunningTargets = runningTargets
	mjson, _ := json.Marshal(cfg)
	mString := string(mjson)

	t.Execute(w, mString)
}

func templateException(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./template/exception.html")
	if err != nil {
		fmt.Fprintln(w, err)
		log4go.Error(err)
		return
	}
	var runningTargets []string
	files := GetRunningFiles()
	for _, val := range files {
		valIp := strings.Split(val, "_")[0]
		runningTargets = append(runningTargets, valIp)
	}
	cfg.RunningTargets = runningTargets
	mjson, _ := json.Marshal(cfg)
	mString := string(mjson)

	t.Execute(w, mString)
}

func templateNetFlow(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./template/netflow.html")
	if err != nil {
		fmt.Fprintln(w, err)
		log4go.Error(err)
		return
	}
	var runningTargets []string
	files := GetRunningFiles()
	for _, val := range files {
		valIp := strings.Split(val, "_")[0]
		runningTargets = append(runningTargets, valIp)
	}
	cfg.RunningTargets = runningTargets
	mjson, _ := json.Marshal(cfg)
	mString := string(mjson)

	t.Execute(w, mString)
}

func templateConfig(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./template/config.html")
	if err != nil {
		fmt.Fprintln(w, err)
		log4go.Error(err)
		return
	}
	var runningTargets []string
	files := GetRunningFiles()
	for _, val := range files {
		valIp := strings.Split(val, "_")[0]
		runningTargets = append(runningTargets, valIp)
	}
	cfg.RunningTargets = runningTargets
	mjson, _ := json.Marshal(cfg)
	mString := string(mjson)

	t.Execute(w, mString)
}

type historyRes struct {
	// 这里有坑，这里的变量名必须大写，引用范围机制
	Data string `json:"data"` //`json:",string"`
}

func indexFunc(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./template/index.html")
	if err != nil {
		fmt.Fprintln(w, err)
		log4go.Error(err)
		return
	}
	var runningTargets []string
	files := GetRunningFiles()
	for _, val := range files {
		valIp := strings.Split(val, "_")[0]
		runningTargets = append(runningTargets, valIp)
	}
	cfg.RunningTargets = runningTargets
	mjson, _ := json.Marshal(cfg)
	mString := string(mjson)

	t.Execute(w, mString)
}

func templateFunc(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./template/tmpl.html")
	if err != nil {
		fmt.Fprintln(w, err)
		log4go.Error(err)
		return
	}
	var runningTargets []string
	files := GetRunningFiles()
	for _, val := range files {
		valIp := strings.Split(val, "_")[0]
		runningTargets = append(runningTargets, valIp)
	}
	cfg.RunningTargets = runningTargets
	mjson, _ := json.Marshal(cfg)
	mString := string(mjson)

	t.Execute(w, mString)
}

//func adminTemplateFunc(w http.ResponseWriter, r *http.Request) {
//	t, err := template.ParseFiles("./template/tmplAdmin.html")
//	if err != nil {
//		fmt.Fprintln(w, err)
//		log4go.Error(err)
//		return
//	}
//	mjson, _ := json.Marshal(cfg)
//	mString := string(mjson)
//
//	t.Execute(w, mString)
//}

func echoFunc(w http.ResponseWriter, r *http.Request) {
	var addr string
	if r.Method == "POST" {
		addr = GetPostArg(r, "addr")
		fmt.Println(addr)
	}

	ret := historyRes{Data: ""}
	ret.Data = httpPostToTarget(addr)

	ret_json, err := json.Marshal(ret)
	if err != nil {
		fmt.Println(err.Error())
	}
	w.Write([]byte(ret_json))
	// 返回时间#数据长度
	//w.Write([]byte(time.Now().String() + "#" + strconv.Itoa(len(body))))
}
func handleConfigRequest(conn *websocket.Conn) {
	var path = "./config/config.cfg"
	// file, _ := os.Open(path)
	// defer file.Close()
	defer conn.Close()
	msg := ""
	err := websocket.Message.Receive(conn, &msg)
	if err != nil {
		fmt.Println("receive config err")
		fmt.Println(err)
		log4go.Error(err)
		return
	}
	if msg == "" {
		return
	}
	if msg == "read" {
		fmt.Println("conf read")

		conf, _ := ioutil.ReadFile(path)
		err = websocket.Message.Send(conn, string(conf))
		if err != nil {
			fmt.Println("conf 发送失败")
			log4go.Error("conf 发送失败")
		}
	} else {
		fmt.Println("conf write")
		err := ioutil.WriteFile(path, []byte(msg), os.ModeAppend)

		if err != nil {
			fmt.Println(err)
			log4go.Error("conf文件写入失败")
		}

	}
}
