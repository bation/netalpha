// speedTest
package main

import (
	"errors"
	"fmt"
	"net"

	"os"
	"time"

	lgg "github.com/AlexStocks/log4go"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// var deviceInfoMap map[string]string

func DeviceSpeed() {
	// brocast.Write("")
	// deviceInfoMap = LoadFile(configFilePath)
	var cfgIP = cfg.Ip
	// var cfgBindwidth = deviceInfoMap["DEVICE_BANDWIDTH"]
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
			go speedTest(name)
		}

	}

}

func speedTest(deviceName string) {
	var (
		downStreamDataSize = 0 // 单位时间内下行的总字节数
		upStreamDataSize   = 0 // 单位时间内上行的总字节数
		// deviceName         = flag.String("i", "eth0", "network interface device name") // 要监控的网卡名称
	)
	fmt.Println("goin speedtest")
	// Find all devices
	// 获取所有网卡
	devices, err := pcap.FindAllDevs()
	if err != nil {
		lgg.Error(err)
	}

	// Find exact device
	// 根据网卡名称从所有网卡中取到精确的网卡
	var device pcap.Interface
	for _, d := range devices {
		if d.Name == deviceName {
			device = d
		}
	}

	// 根据网卡的ipv4地址获取网卡的mac地址，用于后面判断数据包的方向
	var ip = findDeviceIpv4(device)
	macAddr, err := findMacAddrByIp(ip)
	if err != nil {
		lgg.Error(err)
		return // ip为空直接退出
		// panic(err)
	}

	fmt.Printf("Chosen device's IPv4: %s\n", findDeviceIpv4(device))
	fmt.Printf("Chosen device's MAC: %s\n", macAddr)

	// 获取网卡handler，可用于读取或写入数据包
	handle, err := pcap.OpenLive(deviceName, 1024 /*每个数据包读取的最大值*/, true /*是否开启混杂模式*/, 30*time.Second /*读包超时时长*/)
	if err != nil {
		lgg.Error(err)
	}
	defer handle.Close()

	// 开启子线程，每一秒计算一次该秒内的数据包大小平均值，并将下载、上传总量置零
	go monitor(&downStreamDataSize, &upStreamDataSize, ip)

	// 开始抓包
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		if interrupt {
			interruptPool += "speed,"
			fmt.Println("speed out")
			return
		}
		// 只获取以太网帧
		ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
		if ethernetLayer != nil {
			ethernet := ethernetLayer.(*layers.Ethernet)
			// 如果封包的目的MAC是本机则表示是下行的数据包，否则为上行
			if ethernet.DstMAC.String() == macAddr {
				downStreamDataSize += len(packet.Data()) // 统计下行封包总大小
			} else {
				upStreamDataSize += len(packet.Data()) // 统计上行封包总大小
			}
		}
	}
}

// 获取网卡的IPv4地址
func findDeviceIpv4(device pcap.Interface) string {
	for _, addr := range device.Addresses {
		defer func() {
			if err := recover(); err != nil {
				lgg.Error(err)
			}
		}()
		if ipv4 := addr.IP.To4(); ipv4 != nil {
			return ipv4.String()
		}
	}
	return ""
	// panic("device has no IPv4")
}

// 根据网卡的IPv4地址获取MAC地址
// 有此方法是因为gopacket内部未封装获取MAC地址的方法，所以这里通过找到IPv4地址相同的网卡来寻找MAC地址
func findMacAddrByIp(ip string) (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		lgg.Error(err)
	}

	for _, i := range interfaces {
		addrs, err := i.Addrs()
		if err != nil {
			lgg.Error(err)
		}

		for _, addr := range addrs {
			if a, ok := addr.(*net.IPNet); ok {
				if ip == a.IP.String() {
					return i.HardwareAddr.String(), nil
				}
			}
		}
	}
	return "", errors.New(fmt.Sprintf("no device has given ip: %s", ip))
}

// 每一秒计算一次该秒内的数据包大小平均值，并将下载、上传总量置零
func monitor(downStreamDataSize *int, upStreamDataSize *int, ip string) {
	var sec = interval
	for {
		if interrupt {
			interruptPool += "monitor,"

			fmt.Println("speed monitor out")
			return
		}
		speedInfo := make(map[string]int)
		speedInfo["ip"] = StringIpToInt(ip)         // 本地ip
		speedInfo["upload"] = *upStreamDataSize     //本地上传速率
		speedInfo["download"] = *downStreamDataSize //本地下载速率
		speedInfo["duration_sec"] = sec             //间隔毫秒
		bandwidth := int(cfg.Bandwidth)
		speedInfo["bandwidth"] = bandwidth // 本地带宽

		// bandwidth := deviceInfoMap["DEVICE_BANDWIDTH"]
		// localIp := StringIpToInt(ip)
		// speedAndPingMap.Store("localIp", localIp)
		// speedAndPingMap.Store("upload", *upStreamDataSize)
		// speedAndPingMap.Store("download", *downStreamDataSize)
		// speedAndPingMap.Store("bandwidth", strconv.Atoi(bandwidth))

		brocastSpeedAndPing.Write(speedInfo)
		os.Stdout.WriteString("")
		// os.Stdout.WriteString(fmt.Sprintf("\r ip:%s Down:%.2fkb/s \t Up:%.2fkb/s", ip, float32(*downStreamDataSize)/1024/sec, float32(*upStreamDataSize)/1024/sec))
		*downStreamDataSize = 0
		*upStreamDataSize = 0
		time.Sleep(time.Duration(interval * 1000 * 1000 * 1000))

	}
}
