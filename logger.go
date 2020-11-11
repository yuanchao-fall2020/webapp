package main

import (
	"flag"
	"log"
	"os"
)

import (
"go/build"
)

var (
	Log      *log.Logger
)


func init() {
	// set location of log file
	var logpath = build.Default.GOPATH + "/src/chat/logger/info.log"

	flag.Parse()
	var file, err1 = os.Create(logpath)

	if err1 != nil {
		panic(err1)
	}
	Log = log.New(file, "", log.LstdFlags|log.Lshortfile)
	Log.Println("LogFile : " + logpath)
}
