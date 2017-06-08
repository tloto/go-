package main

import (
	"io"
	"log"
	"os"

	configo "gopkg.in/configo.v2"

	"git.g7n3.com/hamster/gautu"
	"git.g7n3.com/hamster/gautu/common"
)

func main() {
	go func() {
		wd, e := os.Getwd()
		path := "." + configo.GetSystemSeparator() + "system.log"
		if e == nil {
			path = wd + configo.GetSystemSeparator() + "system.log"

		}

		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND|os.O_SYNC, os.ModePerm)
		if err != nil {
			common.DoLog("file open error", "gautu")
			return
		}

		w := io.MultiWriter(file, os.Stdout)
		log.SetOutput(w)
		log.SetFlags(log.LstdFlags | log.Llongfile)
		log.Println("server start")
		common.DoLog("server start", "gautu")

	}()

	gautu.Start()

}
