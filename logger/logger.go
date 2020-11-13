package logger

import (
	"flag"
	"log"
	"os"
)

var (
	Log      *log.Logger
)


func init() {
	// set location of log file
	//var logpath = build.Default.GOPATH + "/src/logger/info.log"
	//var logpath = build.Default.GOPATH + "/opt/logs/csye6225.log"
	var logpath = "/opt/logs/csye6225.log"

	flag.Parse()
	var file, err1 = os.Create(logpath)

	if err1 != nil {
		panic(err1)
	}
	Log = log.New(file, "", log.LstdFlags|log.Lshortfile)
	Log.Println("LogFile : " + logpath)
}
