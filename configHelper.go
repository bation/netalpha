package main

import (
	"bufio"
	"fmt"
	lgg "github.com/AlexStocks/log4go"
	"github.com/google/gopacket/pcap"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type Config struct {
	attrMap   map[string]string //配置文件中的键值对
	path      string            //配置文件地址
	Targets   []string          `json:"targets"`   //网关地址
	Bandwidth float64           `json:"bandwidth"` //网卡带宽
	Interval  int               `json:"interval"`  // 间隔时间 单位秒
	Ip        string            `json:"ip"`        // 本地ip
	name      string            //设备名
}

/*
 key:
DEVICE_IP:本地IP DURATION:间隔时间 DEVICE_BANDWIDTH:带宽 DEVICE_TARGET:网关地址，多个以逗号","分割
*/
func (c *Config) GetValueByKey(key string) string {
	_, ok := c.attrMap[key]
	if ok {
		//存在
		return c.attrMap[key]
	}
	return ""
}
func readLine(r *bufio.Reader) (string, error) {
	line, isprefix, err := r.ReadLine()
	for isprefix && err == nil {
		var bs []byte
		bs, isprefix, err = r.ReadLine()
		line = append(line, bs...)
	}
	return string(line), err
}

// key==> DEVICE_IP:本地IP DURATION:间隔时间 DEVICE_BANDWIDTH:带宽 DEVICE_TARGET:网关地址，多个以逗号","分割
// 返回值：bool 表示设置value后是否能正确拿到设备名
// 设置键值对后会重新读取配置文件并赋值给config
func (c *Config) SetValueByKey(key string, value string) (bool, error) {
	iscfgOk := false
	f, err := os.Open(c.path)
	defer f.Close()
	if err != nil {
		fmt.Println("read log file err:")
		fmt.Println(err)
		lgg.Error(err)
		return iscfgOk, err
	}
	sumstr := ""
	reader := bufio.NewReader(f)
	for {
		line, err := readLine(reader)
		if err != nil {
			if err == io.EOF {
				// 文件末尾
				break
			} else {
				return iscfgOk, err
			}
		} else {
			if line == "" {
				continue
			}
			if strings.Contains(line, key) && !strings.HasPrefix(line, "#") {
				sumstr = sumstr + key + " = " + value + "\n"
				continue
			}
			sumstr += line + "\n"
		}
	}
	sumstr = sumstr[0 : len(sumstr)-1]
	err = ioutil.WriteFile(c.path, []byte(sumstr), os.ModeAppend)
	fmt.Println("*************************************************************")
	fmt.Println("修改后的配置文件:")
	fmt.Println(sumstr)
	fmt.Println("*************************************************************")
	iscfgOk = c.Init(c.path) //重新加载配置文件
	return iscfgOk, err
}

// 初始化DEVICE_TARGET
func (c *Config) initValue() {
	strs := strings.Split(c.GetValueByKey("DEVICE_TARGET"), ",")
	var pingTargets []string // []string{"192.168.96.230", "192.168.96.231"}
	for _, val := range strs {
		fmt.Printf("*****ping target: %s ******\n", val)
		pingTargets = append(pingTargets, strings.TrimSpace(val))
	}
	c.Targets = pingTargets
	c.Interval = strToInt(c.GetValueByKey("INTERVAL_SEC"))
	c.Ip = c.GetValueByKey("DEVICE_IP")
	c.Bandwidth = strToFloat64(c.GetValueByKey("DEVICE_BANDWIDTH"))
}

//判断配置文件是否已正确配置
//返回bool true 以正确配置
func (c *Config) IsConfigSetup() bool {
	c.name = ""
	var cfgIP = c.GetValueByKey("DEVICE_IP")
	isSetup := false
	// Find all devices
	devices, err := pcap.FindAllDevs()
	if err != nil {
		lgg.Error(err)
	}
	// Print device information
	//fmt.Println("Devices found:")
	for _, device := range devices {
		var name = device.Name
		var currentDeviceIP = ""

		//fmt.Println("\nName: ", device.Name)
		//fmt.Println("Description: ", device.Description)
		//fmt.Println("Devices addresses: ", device.Description)
		for _, address := range device.Addresses {
			currentDeviceIP = address.IP.String()
			//fmt.Println("- IP address: ", address.IP)
			//fmt.Println("- Subnet mask: ", address.Netmask)
		}
		if currentDeviceIP == cfgIP {
			//go speedTest(name)
			fmt.Println("已找到ip：" + cfgIP + "对应的设备，设备名称：" + name)
			c.name = name
			isSetup = true
		}

	}
	return isSetup
}

// 初始化配置文件
// 返回 是否需要重新配置
func (c *Config) Init(path string) bool {
	//从配置导入文件
	c.path = path
	F, err := os.Open(c.path)
	if err != nil {
		lgg.Error("config.cfg open failed\n")
		panic("config.cfg open failed\n")
	}
	filemap := make(map[string]string)
	bufferReader := bufio.NewReader(F)
	eof := false
	for !eof {
		line, err := bufferReader.ReadString('\n')
		if err == io.EOF {
			err = nil
			eof = true
		} else if err != nil {
			lgg.Error(err)
			panic("parse file error\n")
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		if len(line) > 1 {
			fileconfig := strings.Split(line, "=")
			if len(fileconfig) == 2 {
				filemap[strings.TrimSpace(fileconfig[0])] = strings.TrimSpace(fileconfig[1])
				fmt.Printf("key: %s ,value: %s\n", fileconfig[0], fileconfig[1])
			}
		}
	}
	c.attrMap = filemap
	c.initValue()
	fmt.Println("loadfile finish\n")
	return c.IsConfigSetup()
}
