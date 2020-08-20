package main

import (
	"fmt"
	"github.com/AlexStocks/log4go"
	"io/ioutil"
	"strings"
	"time"
)

func getNewLogger(ip string, minutes string) log4go.Logger {
	tnow := time.Now().Format("2006-01-02-15-04-05")
	filename := ip + "_" + minutes + "m_" + tnow + ".log"
	filePath := "./log/" + filename
	log := log4go.NewLogger()
	log.AddFilter("stdout", log4go.DEBUG, log4go.NewConsoleLogWriter(false))
	flw := log4go.NewFileLogWriter(filePath, true, 0)
	flw.SetFormat("[%D %T]#%M") //("[%D %T] [%L] (%S) %M")
	flw.SetRotate(false)
	log.AddFilter("log", log4go.INFO, flw)
	return log
}

// 是否已经对当前ip运行测试
func isRepeatTask(targets []string) bool {
	isTaskRunning := false
	files := listAllLogFileName(1, "./log")
	for _, ip := range targets {
		for _, fname := range files {
			fmt.Println(fname)
			vls := strings.Split(fname, "_")
			if len(vls) == 3 {
				stt := strings.Split(vls[2], ".")[0]
				min := vls[1]
				logIp := vls[0]
				startTime, _ := time.ParseInLocation("2006-01-02-15-04-05", stt, time.Local)
				endTime := getLogEndTime(min, startTime)
				// 结束时间-当前时间 <=0 时间已过
				deltTime := endTime.Sub(time.Now()).Seconds()
				if deltTime >= 0 && ip == logIp {
					isTaskRunning = true
					break
				}
			}
		}
	}

	return isTaskRunning
}

func GetRunningFiles() []string {
	files := listAllLogFileName(1, "./log")
	var temp []string
	for _, fname := range files {
		fmt.Println(fname)
		vls := strings.Split(fname, "_")
		if len(vls) == 3 {
			stt := strings.Split(vls[2], ".")[0]
			min := vls[1]
			startTime, _ := time.ParseInLocation("2006-01-02-15-04-05", stt, time.Local)
			endTime := getLogEndTime(min, startTime)
			// 结束时间-当前时间 <=0 时间已过
			// deltTime > 0 说明是未来的时间
			deltTime := endTime.Sub(time.Now()).Seconds()
			if deltTime >= 0 {
				temp = append(temp, fname)
			}
		}
	}
	return temp
}

// level 目录级别 单目录1
func listAllLogFileName(level int, fileDir string) []string {
	logFiles := []string{}
	//pathSeparator := string(os.PathSeparator)
	var listFilePrefix string = "  "
	files, dirErr := ioutil.ReadDir(fileDir)
	if dirErr != nil {
		log4go.Error("读取目录失败" + fileDir)
	}
	tmpPrefix := ""
	for i := 1; i < level; i++ {
		tmpPrefix = tmpPrefix + listFilePrefix
	}

	for _, onefile := range files {
		if onefile.IsDir() {
			//fmt.Printf("\033[34m %s %s \033[0m \n" , tmpPrefix, onefile.Name());
			//fmt.Println(tmpPrefix, onefile.Name(), "目录:")
			listAllFileByName(level+1, onefile.Name())
		} else {
			//fmt.Print(onefile.Name()+" ")
			if strings.Contains(onefile.Name(), ".log") {
				logFiles = append(logFiles, onefile.Name())
			}
		}
	}
	return logFiles

}
