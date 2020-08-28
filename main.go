// 需要一个main函数 跑所有线程
package main

import (
	"fmt"
	"github.com/AlexStocks/log4go"
	"os"
	"sync"
)

var mainThread = sync.WaitGroup{}

////整合广播ping speed
//var brocastSpeed = NewBroadcaster()
//var chanelSpeedRcver = brocastSpeed.Listen()
var netUsingQuene Queue // 接收及发送数据的队列--本地网卡流量
var cfg Config
var statusQuene Queue // 接收及发送数据的队列--网关节点状态监控
var stopQuene Queue   // 用于存储需停止发送icmp的IP

var defaultLogger log4go.Logger

func main() {
	// 初始化配置文件
	iscfgOk := cfg.Init("./config/config.cfg")
	//fmt.Println("",cfg.offlineRepURL=="")
	netUsingQuene.Init()
	statusQuene.Init()
	stopQuene.Init()
	lgg := log4go.NewLogger()
	defer lgg.Close()
	//lgg.LoadConfiguration("./config/log4go.xml")
	//lgg.SetAsDefaultLogger()
	//lgg.Info("start running")
	lgg.AddFilter("stdout", log4go.ERROR, log4go.NewConsoleLogWriter(false))
	flw := log4go.NewFileLogWriter("./log/neta.log", true, len(cfg.Targets)*100*2)
	flw.SetFormat("[%D %T]#%M") //("[%D %T] [%L] (%S) %M")
	flw.SetRotate(true)
	flw.SetRotateLines(80000)
	lgg.AddFilter("log", log4go.INFO, flw)
	lgg.SetAsDefaultLogger()
	lgg.Info("start running")
	defaultLogger = lgg
	httpHandle()

	//cfg.setValueByKey("DURATION","5")
	//iscfgOk, _ = cfg.SetValueByKey("DEVICE_IP", " 192.168.96.183")
	// DONE 判断本地是否包含配置文件中的DEVICE_IP，不包含则说明配置文件需要修改。
	fmt.Printf("间隔：%d 秒, ip:%s, 设备名：%s, 带宽：%f Mbps, 配置文件位置：%s, 网关：%s \n", cfg.Interval, cfg.Ip, cfg.name, cfg.Bandwidth, cfg.path, cfg.Targets)
	if !iscfgOk {
		//需要重新修改配置文件
		fmt.Println("请修改配置文件中的 DEVICE_IP 和 DEVICE_BANDWIDTH，然后重新运行！")
		os.Exit(1)
	}
	// 正式开始

	// ping 直接写入log
	for statusQuene.Size() > 0 {
		statusQuene.Dequeue()
	}
	// 网络通断监控
	go GoPing(cfg.Targets, false, &lgg, 0)
	// 网卡流量监控
	go DeviceSpeed()
	/* web服务*/
	go startLiteServer()

	mainThread.Add(1)
	mainThread.Wait()
	//fmt.Println("finish")
	//goto lableStart

}
