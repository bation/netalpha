package main

import (
	log4go "github.com/AlexStocks/log4go"
	"time"
)

func getNewLogger(ip string, minutes string) log4go.Logger {
	tnow := time.Now().Format("2006-01-02-15-04-05")
	fileName := "./log/" + ip + "_" + minutes + "min_" + tnow + ".log"
	log := log4go.NewLogger()
	log.AddFilter("stdout", log4go.DEBUG, log4go.NewConsoleLogWriter(false))
	flw := log4go.NewFileLogWriter(fileName, true, 0)
	flw.SetFormat("[%D %T]#%M") //("[%D %T] [%L] (%S) %M")
	flw.SetRotate(false)
	log.AddFilter("log", log4go.INFO, flw)
	return log
}
