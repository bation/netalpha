package main

import (
	"fmt"
	"github.com/AlexStocks/log4go"
	"net"

	// "strconv"
	"time"
)

var isStandalone = false

const (
	ONLINE      = "ONLINE"
	OFFLINE     = "OFFLINE"
	WARN        = "LOSTRATEG1"   // 丢包率超过1% 警告
	HIGHLATENCY = "HIGH_LATENCY" // 平均延迟超过300ms
)
const NOTIME = "2006-01-02 15:04:05"

//	获取 通断 丢包率 抖动
// args 需要ping的地址
// min 执行分钟数 0代表一直运行
func GoPing(args []string, logger *log4go.Logger, min int) {
	if len(args) == 1 {
		isStandalone = true
	}
	var now = time.Now() // 结束时间
	var endTime time.Time
	remainingMin, _ := time.ParseDuration(intToStr(min) + "m")
	endTime = now.Add(remainingMin)
	if now == endTime {
		endTime, _ = time.ParseInLocation(NOTIME, NOTIME, time.Local)
	}

	var count int      //要发送的回显请求数。
	var timeout int64  //等待每次回复的超时时间(毫秒)
	var size int       //要发送缓冲区大小。
	var neverstop bool //Ping 指定的主机，直到停止。
	count = 4
	timeout = 3000
	size = 32
	neverstop = true

	argsmap := map[string]interface{}{}

	argsmap["w"] = timeout
	argsmap["n"] = count
	argsmap["l"] = size
	argsmap["t"] = neverstop

	for _, host := range args {
		if host == "" {
			continue
		}
		go ping(host, argsmap, logger, endTime)
	}

}

func ping(host string, args map[string]interface{}, logger *log4go.Logger, endTime time.Time) {
	var count int
	var size int
	var timeout int64
	var neverstop bool
	count = args["n"].(int)
	size = args["l"].(int)
	timeout = args["w"].(int64)
	neverstop = args["t"].(bool)

	cname, _ := net.LookupCNAME(host)
	starttime := time.Now()
	conn, err := net.DialTimeout("ip4:icmp", host, time.Duration(timeout*1000*1000))
	ip := conn.RemoteAddr()
	fmt.Println("正在 Ping " + cname + " [" + ip.String() + "] 具有 32 字节的数据:")

	var seq int64 = 1
	id0, id1 := genidentifier(host)
	const ECHO_REQUEST_HEAD_LEN = 8

	sendN := int64(0)
	recvN := int64(0)
	lostN := int64(0)
	shortT := int64(-1)
	longT := int64(-1)
	sumT := int64(0)
	var quene Queue
	quene.Init()
	unLimitedTime, _ := time.ParseInLocation(NOTIME, NOTIME, time.Local)
	var isUnlimited = false
	if unLimitedTime == endTime {
		// 不限时间
		isUnlimited = true
	}

	for count > 0 || neverstop {
		if interrupt {
			fmt.Println("ping out")
			interruptPool += "ping,"
			return
		}
		if !isUnlimited {
			if endTime.Sub(time.Now()).Seconds() <= 0 {
				// 时间到，结束
				logger.Info("*****因到达运行时间而终止*****")
				return
			}
		}

		sendN++
		var msg []byte = make([]byte, size+ECHO_REQUEST_HEAD_LEN)
		msg[0] = 8                        // echo
		msg[1] = 0                        // code 0
		msg[2] = 0                        // checksum
		msg[3] = 0                        // checksum
		msg[4], msg[5] = id0, id1         //identifier[0] identifier[1]
		msg[6], msg[7] = gensequence(seq) //sequence[0], sequence[1]

		length := size + ECHO_REQUEST_HEAD_LEN

		check := checkSum(msg[0:length])
		msg[2] = byte(check >> 8)
		msg[3] = byte(check & 255)

		conn, err = net.DialTimeout("ip:icmp", host, time.Duration(timeout*1000*1000))
		if err != nil {
			logger.Error("icmp 建立连接失败，重试：" + err.Error())
			continue
		}

		starttime = time.Now()
		conn.SetDeadline(starttime.Add(time.Duration(timeout * 1000 * 1000)))
		_, err = conn.Write(msg[0:length])

		const ECHO_REPLY_HEAD_LEN = 20

		var receive []byte = make([]byte, ECHO_REPLY_HEAD_LEN+length)
		n, err := conn.Read(receive)
		_ = n

		var endduration int64 = int64(time.Since(starttime)) / (1000 * 1000)

		sumT += endduration
		// 发送频次
		if isStandalone {
			// 异常节点监测，每5秒一次，不可调
			time.Sleep(time.Duration(5 * 1000 * 1000 * 1000))
		} else {
			time.Sleep(time.Duration(interval * 1000 * 1000 * 1000))

		}

		if err != nil || receive[ECHO_REPLY_HEAD_LEN+4] != msg[4] || receive[ECHO_REPLY_HEAD_LEN+5] != msg[5] || receive[ECHO_REPLY_HEAD_LEN+6] != msg[6] || receive[ECHO_REPLY_HEAD_LEN+7] != msg[7] || endduration >= int64(timeout) || receive[ECHO_REPLY_HEAD_LEN] == 11 {
			lostN++
			// fmt.Println("对 " + cname + "[" + ip.String() + "]" + " 的请求超时。")
		} else {
			if shortT == -1 {
				shortT = endduration
			} else if shortT > endduration {
				shortT = endduration
			}
			if longT == -1 {
				longT = endduration
			} else if longT < endduration {
				longT = endduration
			}
			recvN++
			// ttl := int(receive[8])
			//			fmt.Println(ttl)
			// fmt.Println("来自 " + cname + "[" + ip.String() + "]" + " 的回复: 字节=32 时间=" + strconv.Itoa(endduration) + "ms TTL=" + strconv.Itoa(ttl))
		}

		seq++
		count--
		quene.Enqueue(endduration)
		if quene.Size() > 3 {
			quene.Dequeue() //.(int64)
		}
		if isStandalone {
			statStandalone(logger, ip.String(), sendN, lostN, recvN, shortT, longT, sumT, endduration, quene)
		} else {
			stat(logger, ip.String(), sendN, lostN, recvN, shortT, longT, sumT, endduration, quene)
		}
	}

}

func checkSum(msg []byte) uint16 {
	sum := 0

	length := len(msg)
	for i := 0; i < length-1; i += 2 {
		sum += int(msg[i])*256 + int(msg[i+1])
	}
	if length%2 == 1 {
		sum += int(msg[length-1]) * 256 // notice here, why *256?
	}

	sum = (sum >> 16) + (sum & 0xffff)
	sum += (sum >> 16)
	var answer uint16 = uint16(^sum)
	return answer
}

func gensequence(v int64) (byte, byte) {
	ret1 := byte(v >> 8)
	ret2 := byte(v & 255)
	return ret1, ret2
}

func genidentifier(host string) (byte, byte) {
	return host[0], host[1]
}

type Status struct {
	Ip          string  `json:"ip"`
	Send        int64   `json:"send"`
	Recv        int64   `json:"recv"`
	Lost        float64 `json:"lost"`
	Duration    int64   `json:"duration"`
	MaxDuration int64   `json:"maxDuration"`
	MinDuration int64   `json:"minDuration"`
	SumDuration int64   `json:"sumDuration"`
}
type StatusStandalone struct {
	Ip   string `json:"ip"`
	Stat string `json:"status"`
	//LostRate float64 `json:"lost"`
}

func stat(logger *log4go.Logger, ip string, sendN int64, lostN int64, recvN int64, shortT int64, longT int64, sumT int64, endduration int64, quene Queue) {
	sumQ := int64(0)
	for i := uint(0); i < quene.Size(); i++ {
		sumQ += quene.list.Get(i).Data.(int64)
	}
	sumAVG := sumQ / int64(quene.Size())
	//fmt.Printf("sumAvg:%d", sumAVG)
	var stat StatusStandalone
	stat.Ip = ip
	ds := endduration
	lost := (float64(lostN) / float64(sendN)) * float64(100)
	//stat.LostRate = lost
	// 状态优先级 ONLINE OFFLINE > HIGHLATENCY WARN
	stat.Stat = ONLINE
	if ds > 300 && ds < 3000 {
		stat.Stat = HIGHLATENCY
	}
	if lost >= 1 && sendN >= 100 {
		stat.Stat = WARN
	}
	if sumAVG >= 3000 {
		stat.Stat = OFFLINE
	}
	//直接写日志，不用Chanel
	logger.Info(structToJsonsting(stat))
	//brocastSpeedAndPing.Write(structToJsonsting(stat))

}
func statStandalone(logger *log4go.Logger, ip string, sendN int64, lostN int64, recvN int64, shortT int64, longT int64, sumT int64, endduration int64, quene Queue) {
	// fmt.Println()
	// fmt.Println(ip, " 的 Ping 统计信息:")
	// fmt.Printf("    数据包: 已发送 = %d，已接收 = %d，丢失 = %d (%d%% 丢失)，\n", sendN, recvN, lostN, int(lostN*100/sendN))
	// fmt.Println("往返行程的估计时间(以毫秒为单位):")
	// if recvN != 0 {
	// 	fmt.Printf("    最短 = %dms，最长 = %dms，平均 = %dms\n", shortT, longT, sumT/sendN)
	// }
	//pingInfo := make(map[string]int)
	//pingInfo["ip"] = StringIpToInt(ip)
	//pingInfo["send"] = sendN                             //发送数量
	//pingInfo["lost"] = lostN                             //丢失数量
	//pingInfo["recv"] = recvN                             //接收数量
	//pingInfo["short"] = shortT                           //最短用时
	//pingInfo["long"] = longT                             // 最长用时
	//pingInfo["lostPercent"] = int(lostN * 100 / sendN) // 丢包率
	//pingInfo["duration"] = endduration                   //发包用时
	var stat Status
	stat.Send = sendN
	stat.Duration = endduration
	stat.Ip = ip
	stat.Lost = (float64(lostN) / float64(sendN)) * float64(100)
	stat.MaxDuration = longT
	stat.MinDuration = shortT
	stat.Recv = recvN
	stat.SumDuration = sumT / sendN
	//直接写日志，不用Chanel
	logger.Info(structToJsonsting(stat))
	//brocastSpeedAndPing.Write(structToJsonsting(stat))
}
