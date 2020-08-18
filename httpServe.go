package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/websocket"

	"github.com/AlexStocks/log4go"
)

//TODO  写日志
func writeSpeedAndPingLog() {
	writeCount := 0
	//些日子
	for {
		writeCount += 1
		if interrupt {
			interruptPool += "write,"
			writeCount = 0
			fmt.Println("write log out")
			return
		}
		if writeCount >= 50000 {
			// 超过5w行 休眠一秒
			time.Sleep(1 * time.Second)
			writeCount = 0
		}
		info := chanelSpeedAndPingRcver.Read().(string)
		if interrupt {
			interruptPool += "write,"
			writeCount = 0
			fmt.Println("write log out")
			return
		}

		log4go.Info(info)
	}
}
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
	// //读取log
lableReachLineMax:
	f, err := os.Open("./log/neta.log")
	defer f.Close()
	if err != nil {
		fmt.Println("read log file err:")
		fmt.Println(err)
		log4go.Error(err)
		return
	}
	reader := bufio.NewReader(f)
	isEOFOnce := false
	fmt.Println("read log now")
	lineCount := 0
	for {
		if interrupt {
			fmt.Println("conn out")
			interruptPool += "conn,"
			return
		}
		line, err := readTheFuckingLine(reader)
		if err != nil {
			if err == io.EOF {
				// 文件末尾
				isEOFOnce = true // 读到了文件末尾
				// 每次到文件末尾休息1秒
				time.Sleep(1 * time.Second)
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
					fmt.Println(err)
					log4go.Error(err)

					break
				}
			}
			lineCount += 1 //行计数
			if lineCount >= (50000) {
				fmt.Println("文件到达日志最大行，准备重新读取文件")
				log4go.Info("文件到达日志最大行，准备重新读取文件")
				ferr := f.Close() // 不然会一直占用io导致日志无法rotate
				if ferr != nil {
					fmt.Println(ferr)
				}
				time.Sleep(1 * time.Second) // 只能休眠<3秒否则icmp可能连接失败
				goto lableReachLineMax
			}

		}
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
func restartProgram(w http.ResponseWriter, r *http.Request) {
	// 接收重启信号
	restartNow()
}
func restartNow() {
	// restart
	interrupt = true
	time.Sleep(6 * time.Second)
	// 计算ping+device(stat+monitor) ，+write, conn
	connCount := 1
	pingCount := len(cfg.Targets)
	deviceThread := 2
	writeThread := 1
	threadCount := pingCount + deviceThread + writeThread + connCount
	interruptedCount := len(strings.Split(interruptPool, ","))
	loopCount := 0
	for {
		loopCount += 1
		if loopCount > 3 {
			fmt.Println("超过时间，强制重启")
			break
		}
		if threadCount <= interruptedCount {
			mainThread.Done()
			reNewBrocaster() // 新建广播
			fmt.Println("restarting now")
			interruptPool = ""
			break
		}
		time.Sleep(1 * time.Second)
	}

}
func handleJsonIntToIP(mString string, ip int) string {
	strArr := strings.Split(mString, ",")
	sumStr := ""
	for index, val := range strArr {
		if strings.Contains(val, "ip") {
			if index == 0 {
				sumStr += "{"
			}
			sumStr += ("\"ip\":\"" + IpIntToString(ip) + "\",")
			continue
		}
		sumStr += (val + ",")
	}
	sumStr = sumStr[0 : len(sumStr)-1]
	return sumStr
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

func startPingTargetsFunc(w http.ResponseWriter, r *http.Request) {
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
		return
	}
	// 开始ping
	for _, ip := range data.Targets {
		fmt.Println(ip + "|||" + intToStr(data.Min))
		// 获得logger后，需要ping target，每个ping和logger都是独立的.
		// 修改ping，加个logger参数？
		var ipLogger = getNewLogger(ip, intToStr(data.Min))
		// 开始独立记录
		go GoPing([]string{ip}, &ipLogger, data.Min)

	}

	/// 以下为结果返回
	endUrl := "/config"
	type result struct {
		Data string `json:"data"`
	}
	var ret result
	ret.Data = endUrl

	ret_json, err := json.Marshal(ret)
	if err != nil {
		log4go.Error("res convert to json failed " + err.Error())
		return
	}
	w.Write([]byte(ret_json))
	return
}
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

func startLiteServer() {
	server = http.Server{
		Addr: ":8769",
	}
	fmt.Println("*********server addr********************")
	fmt.Println("server.Addr: " + server.Addr)
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println("服务启动失败" + err.Error())
	}
}
func httpHandle() {
	// http.Handle("/css/", http.FileServer(http.Dir("template")))
	http.HandleFunc("/startPingTargets", startPingTargetsFunc) // 开始ping选择的地址

	http.HandleFunc("/echo", echoFunc)                             // 传输数据 读写
	http.Handle("/config", websocket.Handler(handleConfigRequest)) // 配置 读写
	http.Handle("/js/", http.FileServer(http.Dir("template")))
	http.Handle("/info", websocket.Handler(handleInfo)) // 只写
	http.HandleFunc("/sendRestart", restartProgram)     //只读
	// http.HandleFunc("/", templateFunc)
	http.HandleFunc("/getHistory", getHistoryFunc)
	http.HandleFunc("/", templateFunc)
	httpAdminHandle()
}
func httpAdminHandle() {
	http.HandleFunc("/admin", adminTemplateFunc)
}
func reNewBrocaster() {
	brocastSpeedAndPing = NewBroadcaster()
	chanelSpeedAndPingRcver = brocastSpeedAndPing.Listen()
}
func getHistoryFunc(w http.ResponseWriter, r *http.Request) {
	var d1 string
	var d2 string
	if r.Method == "POST" {
		d1 = GetPostArg(r, "d1")
		d2 = GetPostArg(r, "d2")
	}
	fmt.Println(d1 + "|||" + d2)
	//body, _ := ioutil.ReadAll(r.Body)
	//fmt.Println(string(body))
	//data := strings.Split(string(body), "&")
	//ds1 := strings.Split(data[0], "=")
	//ds2 := strings.Split(data[1], "=")
	//d1 := ds1[1]
	//d2 := ds2[1]
	res := getHistroy(d1, d2)
	ret := historyRes{Data: ""}
	ret.Data = "{\"result\":\"" + res + "\"}"
	//fmt.Println(ret.Result)
	if len(res) > 100 {
		fmt.Println(ret.Data[0:100] + "……")
	}
	if res == "" {
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

type historyRes struct {
	// 这里有坑，这里的变量名必须大写，引用范围机制
	Data string `json:"data"` //`json:",string"`
}

func templateFunc(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./template/tmpl.html")
	if err != nil {
		fmt.Fprintln(w, err)
		log4go.Error(err)
		return
	}

	mjson, _ := json.Marshal(cfg)
	mString := string(mjson)

	t.Execute(w, mString)
}
func adminTemplateFunc(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./template/tmplAdmin.html")
	if err != nil {
		fmt.Fprintln(w, err)
		log4go.Error(err)
		return
	}
	mjson, _ := json.Marshal(cfg)
	mString := string(mjson)

	t.Execute(w, mString)
}
