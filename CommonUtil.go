package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/AlexStocks/log4go"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
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

type PostMapData struct {
	Url           string
	Data          map[string]string //post要传输的数据，必须key value必须都是string
	DataInterface map[string]interface{}
}

//适用于 application/x-www-form-urlencoded
func (r *PostMapData) PostWithAppEncoded() ([]byte, error) {
	client := &http.Client{}
	//post要提交的数据
	DataUrlVal := url.Values{}
	for key, val := range r.Data {
		DataUrlVal.Add(key, val)
	}
	req, err := http.NewRequest("POST", r.Url, strings.NewReader(DataUrlVal.Encode()))
	if err != nil {
		return nil, err
	}
	//伪装头部
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.8,en-US;q=0.6,en;q=0.4")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Length", "25")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Cookie", "")
	//req.Header.Add("Host","www.lagou.com")
	//req.Header.Add("Origin","https://www.lagou.com")
	//req.Header.Add("Referer","https://www.lagou.com/jobs/list_python?labelWords=&fromSearch=true&suginput=")
	req.Header.Add("X-Anit-Forge-Code", "0")
	req.Header.Add("X-Anit-Forge-Token", "None")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	//提交请求
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	//读取返回值
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return result, nil
}
func (r *PostMapData) PostWithFormData() ([]byte, error) {
	body := new(bytes.Buffer)
	w := multipart.NewWriter(body)
	for k, v := range r.Data {
		w.WriteField(k, v)
	}
	w.Close()
	req, err := http.NewRequest(http.MethodPost, r.Url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	fmt.Println(resp.StatusCode)
	fmt.Printf("%s", data)
	return data, nil
}

// 发送GET请求
// url：         请求地址
// response：    请求返回的内容
func Get(url string) string {

	// 超时时间：5秒
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		panic(err)
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

	return result.String()
}

// 发送POST请求
// url：         请求地址
// data：        POST请求提交的数据
// contentType： 请求体格式，如：application/json
// content：     请求放回的内容
func Post(url string, data interface{}, contentType string) string {

	// 超时时间：5秒
	client := &http.Client{Timeout: 5 * time.Second}
	jsonStr, _ := json.Marshal(data)
	resp, err := client.Post(url, contentType, bytes.NewBuffer(jsonStr))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	result, _ := ioutil.ReadAll(resp.Body)
	return string(result)
}

// 允许跨域 CRS
func AllowCrossOrigin(w http.ResponseWriter) http.ResponseWriter {
	w.Header().Set("Access-Control-Allow-Origin", "*")          //允许访问所有域
	w.Header().Set("content-type", "application/json")          //返回数据格式是json
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET") //允许的请求方法
	w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, "+
		"WG-App-Version, WG-Device-Id, WG-Network-Type, WG-Vendor, WG-OS-Type, WG-OS-Version, WG-Device-Model, WG-CPU, WG-Sid, WG-App-Id, WG-Token") //header的类型
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	return w
}

// 取最小值 i := Minimum(1, 3, 5, 7, 9, 10, -1, 1).(int)
//    fmt.Printf("i = %d\n", i)
func Minimum(first interface{}, rest ...interface{}) interface{} {
	minimum := first

	for _, v := range rest {
		switch v.(type) {
		case int:
			if v := v.(int); v < minimum.(int) {
				minimum = v
			}
		case float64:
			if v := v.(float64); v < minimum.(float64) {
				minimum = v
			}
		case string:
			if v := v.(string); v < minimum.(string) {
				minimum = v
			}
		}
	}
	return minimum
}

// 取最小值 i := Minimum(1, 3, 5, 7, 9, 10, -1, 1).(int)
//    fmt.Printf("i = %d\n", i)
func Minimum1(rest []float64) float64 {
	minimum := rest[0]

	for _, v := range rest {
		if v < minimum {
			minimum = v
		}
	}
	return minimum
}
