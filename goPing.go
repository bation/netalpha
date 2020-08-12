package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	lgg "github.com/AlexStocks/log4go"

	// "strconv"
	"time"
)

func GoPing(args []string) {
	var count int      //要发送的回显请求数。
	var timeout int64  //等待每次回复的超时时间(毫秒)
	var size int       //要发送缓冲区大小。
	var neverstop bool //Ping 指定的主机，直到停止。
	count = 4
	timeout = 3000
	size = 32
	neverstop = true

	// flag.Int64Var(&timeout, "w", 3000, "等待每次回复的超时时间(毫秒)。")
	// flag.IntVar(&count, "n", 4, "要发送的回显请求数。")
	// flag.IntVar(&size, "l", 32, "要发送缓冲区大小。")
	// flag.BoolVar(&neverstop, "t", true, "Ping 指定的主机，直到停止。")

	// flag.Parse()
	// args := flag.Args() // 从命令行接收改为从形参接收
	fmt.Println("args: ", args)

	if len(args) < 1 {
		fmt.Println("Usage: ", os.Args[0], "host")
		flag.PrintDefaults()
		flag.Usage()
		os.Exit(1)
	}

	ch := make(chan int)
	argsmap := map[string]interface{}{}

	argsmap["w"] = timeout
	argsmap["n"] = count
	argsmap["l"] = size
	argsmap["t"] = neverstop

	for _, host := range args {
		if host == "" {
			continue
		}
		go ping(host, ch, argsmap)
	}

	for i := 0; i < len(args); i++ {
		<-ch
	}

	os.Exit(0)
}

func ping(host string, c chan int, args map[string]interface{}) {
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

	for count > 0 || neverstop {
		if interrupt {
			fmt.Println("ping out")
			interruptPool += "ping,"
			return
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
			fmt.Println("icmp 建立连接失败，本线程终止：" + err.Error())
			lgg.Error("icmp 建立连接失败，本线程终止：" + err.Error())
			break
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

		time.Sleep(time.Duration(interval * 1000 * 1000 * 1000))

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
		stat(ip.String(), sendN, lostN, recvN, shortT, longT, sumT, endduration)
	}

	c <- 1
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

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		lgg.Error(err)
		os.Exit(1)
	}
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
	Ip string `json:"ip"`
	Send int64 `json:"send"`
	Recv int64 `json:"recv"`
	Lost float64 `json:"lost"`
	Duration int64 `json:"duration"`
	MaxDuration int64 `json:"maxDuration"`
	MinDuration int64 `json:"minDuration"`
	SumDuration int64 `json:"sumDuration"`
}
func stat(ip string, sendN int64, lostN int64, recvN int64, shortT int64, longT int64, sumT int64, endduration int64) {
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
	stat.Lost = (float64(lostN)/float64(sendN))*100
	stat.MaxDuration = longT
	stat.MinDuration = shortT
	stat.Recv = recvN
	stat.SumDuration = sumT/sendN

	brocastSpeedAndPing.Write(structToJsonsting(stat))
}
