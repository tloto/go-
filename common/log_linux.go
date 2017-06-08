package common

import (
	"log"
	"log/syslog"

	"strings"

	"gopkg.in/configo.v2"
)

func DoLog(c string, cate string) {
	p, e := configo.Get("log")
	if e != nil {
		log.Println("dolog error")
		return
	}
	addr := p.MustGet("addr", "10.162.90.117")
	port := p.MustGet("port", "80")
	remote := strings.Join([]string{addr, port}, ":")
	sysLog, err := syslog.Dial("tcp", remote, syslog.LOG_INFO, cate)
	if err != nil {
		log.Fatal(err)
	}
	defer sysLog.Close()
	sysLog.Info(c)
}
