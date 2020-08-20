// 需要一个main函数 跑所有线程
package main

import (
	"fmt"
	"github.com/AlexStocks/log4go"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var mainThread = sync.WaitGroup{}
var server http.Server
var interrupt bool
var interruptPool string
var interval = 5 //间隔 秒
////整合广播ping speed
//var brocastSpeed = NewBroadcaster()
//var chanelSpeedRcver = brocastSpeed.Listen()
var netUsingQuene Queue // 接收及发送数据的队列--本地网卡流量
var cfg Config
var statusQuene Queue // 接收及发送数据的队列--网关节点状态监控

func main() {
	// 初始化配置文件
	iscfgOk := cfg.Init("./config/config.cfg")
	netUsingQuene.Init()
	statusQuene.Init()
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

	interval = cfg.Interval
	// ping 直接写入log
	for statusQuene.Size() > 0 {
		statusQuene.Dequeue()
	}
	go GoPing(cfg.Targets, false, &lgg, 0)

	/* web服务*/
	go startLiteServer()
	//ll := getNewLogger("192.168.96.230","60")
	//for{
	//	ll.Info("asldfjlasdlfj")
	//	time.Sleep(1*time.Second)
	//}

	//lableStart:
	//interrupt = false
	//
	//time.Sleep(1 * time.Second)
	//// 删除日志
	//beforeRestartDelLog()
	//
	///**
	//获取 通断 丢包率 抖动
	//*/
	//go GoPing(pingTargets)
	///**
	//通过配置文件获取本地网卡的上行和下行速率
	//*/
	//go DeviceSpeed()

	mainThread.Add(1)
	mainThread.Wait()
	//fmt.Println("finish")
	//goto lableStart

}
func beforeRestartDelLog() {
	log4go.Close()
	// 每次运行删除日志
	oserr := os.Rename("./log/neta.log", "./log/neta.log"+strconv.FormatInt(time.Now().Unix(), 10))
	if oserr != nil {
		fmt.Println("重命名日志失败：")
		fmt.Println(oserr)
		log4go.Error("删除日志失败：" + oserr.Error())
	}
}
